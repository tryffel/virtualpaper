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
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/models"
)

type StatsStore struct {
	db *sqlx.DB
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
where user_id = $1
order by updated_at desc limit 10;
`

	err = s.db.Select(&stats.LastDocumentsUpdated, sql, userId)
	return stats, s.parseError(err, "get last updated docs")
}

func (s *StatsStore) GetSystemStats() (*models.SystemStatistics, error) {

	sql := `
SELECT
    count(distinct(d.id)) AS documents_total,
    ( 
        select sum(d.size) from documents d
    ) as documents_size,
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
	return stats, s.parseError(err, "get system stats")
}
