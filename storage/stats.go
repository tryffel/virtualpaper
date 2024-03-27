/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type StatsStore struct {
	db    *sqlx.DB
	cache *cache.Cache
}

func NewStatsStore(db *sqlx.DB) *StatsStore {
	return &StatsStore{
		db:    db,
		cache: cache.New(30*time.Second, time.Minute),
	}
}

func (s *StatsStore) Name() string {
	return "Stats"
}

func (s *StatsStore) parseError(e error, action string) error {
	return getDatabaseError(e, s, action)
}

func (s *StatsStore) GetUserDocumentStats(userId int) (*models.UserDocumentStatistics, error) {
	sql := `

-- document count
select count(id), 1 as ordering from documents where user_id = $1
union

-- metadata key count
select count(id), 2 from metadata_keys where user_id = $1
union

select count(id), 3 from metadata_values where user_id = $1
order by ordering asc;
`

	stats := &models.UserDocumentStatistics{}

	values := []*int{&stats.NumDocuments, &stats.NumMetadataKeys, &stats.NumMetadataValues}

	rows, err := s.db.Query(sql, userId)
	if err != nil {
		return stats, s.parseError(err, "get document counts")
	}
	defer rows.Close()

	var order int

	i := 0
	for rows.Next() {
		err = rows.Scan((values)[i], &order)

		if err != nil {
			logrus.Errorf("scan int: %v", err)
		} else {
			i += 1
		}
	}
	rows.Close()

	sql = `
select 
	extract(year from date) as year, 
	count(id) as count
from documents
where user_id = $1
group by extract(year from date)
order by year desc;
`

	err = s.db.Select(&stats.YearlyStats, sql, userId)
	if err != nil {
		return stats, s.parseError(err, "get yearly stats")
	}

	sql = `
select id
from documents
where user_id = $1 AND deleted_at IS NULL
order by updated_at desc limit 10;
`

	err = s.db.Select(&stats.LastDocumentsUpdated, sql, userId)

	sql = `
select id
from documents
where user_id = $1 AND deleted_at IS NULL
order by created_at desc limit 10;
`

	err = s.db.Select(&stats.LastDocumentsAdded, sql, userId)

	sql = `
select id
from documents
where user_id = $1 AND deleted_at IS NULL AND favorite=true
order by date desc limit 50;
`

	err = s.db.Select(&stats.Favorites, sql, userId)

	sql = `
select t.document_id as document_id
from (
         select document_id, created_at,
                row_number() over (
                    partition by document_id
                    order by created_at desc
                    ) as rn
         from document_view_history
         where user_id = $1) t
LEFT JOIN documents ON t.document_id = documents.id
where rn =1 AND documents.deleted_at IS NULL
order by t.created_at desc
limit 10;
`
	err = s.db.Select(&stats.LastDocumentsViewed, sql, userId)
	return stats, s.parseError(err, "get last updated docs")
}

func (s *StatsStore) GetSystemStats() (*models.SystemStatistics, error) {

	cacheKey := "system-stats"

	cached, ok := s.cache.Get(cacheKey)
	if ok {
		cachedStats, ok := cached.(*models.SystemStatistics)
		if ok {
			logrus.Debugf("storage.GetSystemStats() using cached result")
			return cachedStats, nil
		} else {
			s.cache.Delete(cacheKey)
		}
	}

	sql := `
SELECT
    count(distinct(d.id)) AS documents_total,
    count(distinct(pq.document_id)) as documents_queued,
    (
        select count(distinct(j.document_id)) as documents_processed_today
        from jobs j
        where date(j.started_at) = date(now())
    ) as documents_processed_today,
    (
        select count(distinct(j.document_id)) as documents_processed_today
        from jobs j
        where date(j.started_at) > date(now()) - interval '1 week'
    ) as documents_processed_past_week,
    (
        select count(distinct(j.document_id)) as documents_processed_today
        from jobs j
        where date(j.started_at) > date(now()) - interval '1 month'
    ) as documents_processed_past_month

from documents d
         left join process_queue pq on d.id = pq.document_id
         left join jobs j on d.id = j.document_id;
`
	stats := &models.SystemStatistics{}

	err := s.db.Get(stats, sql)
	if err != nil {
		return stats, s.parseError(err, "get system stats")
	}

	totalSize, err := getStorageTotalSize()
	if err != nil {
		e := errors.ErrInternalError
		e.Err = fmt.Errorf("get total storage size: %v", err)
	}

	stats.DocumentsTotalSize = int64(totalSize)
	stats.DocumentsTotalSizeString = models.GetPrettySize(stats.DocumentsTotalSize)

	s.cache.SetDefault(cacheKey, stats)
	return stats, nil
}

func getStorageTotalSize() (uint64, error) {
	logrus.Warningf("start calculating storage total size, this could take a while")
	path := config.C.Processing.DocumentsDir
	var size uint64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return err
	})

	logrus.Infof("total storage size")
	return size, err
}
