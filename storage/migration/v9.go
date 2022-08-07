/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
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

const schemaV9 = `
CREATE TABLE document_history (
	id SERIAL PRIMARY KEY,
	document_id TEXT NOT NULL,
	action TEXT NOT NULL,
	old_value TEXT NOT NULL DEFAULT '',
	new_value TEXT NOT NULL DEFAULT '',
	user_id INTEGER,
	created_at TIMESTAMPTZ default now(),
	
	CONSTRAINT fk_document 
		FOREIGN KEY (document_id) 
		REFERENCES documents(id) 
		ON DELETE CASCADE,

	CONSTRAINT fk_user
		FOREIGN KEY (user_id) 
		REFERENCES users(id) 
);

`
