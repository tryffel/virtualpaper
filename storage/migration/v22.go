/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2024  Tero Vierimaa
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

const schemaV22 = `
CREATE TABLE properties (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT '',
    global BOOL NOT NULL DEFAULT FALSE,
    is_unique BOOL NOT NULL DEFAULT FALSE,
    is_exclusive BOOL NOT NULL DEFAULT FALSE,
    counter INT NOT NULL DEFAULT 0,
    counter_offset INT NOT NULL DEFAULT 0,
    prefix TEXT NOT NULL DEFAULT '',
    mode TEXT NOT NULL DEFAULT '',
    read_only bool NOT NULL DEFAULT FALSE,
    date_fmt TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
                        
	CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT c_user_id_name_unique UNIQUE (user_id, name)
);

CREATE UNIQUE INDEX properties_c_global_name_unique on properties (name) WHERE global IS NULL;

CREATE TABLE document_properties (
    id SERIAL PRIMARY KEY,
    document_id TEXT NOT NULL,
    property_id INT NOT NULL,
    user_id INT NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    is_unique BOOL NOT NULL DEFAULT FALSE,
    is_exclusive BOOL NOT NULL DEFAULT FALSE,
    global BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
   
	CONSTRAINT fk_user_id FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_document_id FOREIGN KEY(document_id) REFERENCES documents(id) ON DELETE CASCADE,
	CONSTRAINT fk_property_id FOREIGN KEY(property_id) REFERENCES properties(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX document_properties_c_exclusive on document_properties (document_id, property_id, user_id) WHERE is_exclusive IS TRUE;
CREATE UNIQUE INDEX document_properties_c_unique on document_properties (property_id, user_id, value) WHERE is_unique IS TRUE;
CREATE UNIQUE INDEX document_properties_c_global_unique on document_properties (property_id, value) WHERE global IS TRUE;
`
