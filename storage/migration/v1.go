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

package migration

const schemaV1 = `

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    email TEXT,
    password TEXT,
	active BOOLEAN DEFAULT true,
	admin BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    user_id INT,
    name TEXT,
    content TEXT,
    filename TEXT UNIQUE,
	hash TEXT NOT NULL DEFAULT '',

	indexed_at TIMESTAMPTZ NOT NULL DEFAULT TO_TIMESTAMP(0),
	awaits_indexing BOOL NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ,

    CONSTRAINT fk_user FOREIGN KEY(user_id)
        REFERENCES users(id)
);

CREATE TABLE jobs (
	id SERIAL PRIMARY KEY,
	document_id INT,
	status INT,
	message TEXT,

	started_at TIMESTAMPTZ DEFAULT now(),
    stopped_at TIMESTAMPTZ,

	CONSTRAINT fk_document FOREIGN KEY(document_id)
        REFERENCES documents(id)
);

`
