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
		return stats, getDatabaseError(err, "stats", "get document counts")
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
		return stats, getDatabaseError(err, "stats", "get yearly stats")
	}

	sql = `
select id
from documents
where user_id = $1
order by updated_at desc limit 10;
`

	err = s.db.Select(&stats.LastDocumentsUpdated, sql, userId)
	return stats, getDatabaseError(err, "stats", "get last updated docs")
}
