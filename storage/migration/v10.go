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

const schemaV10 = `
CREATE TABLE linked_documents (
	doc_a_id TEXT NOT NULL,
	doc_b_id TEXT NOT NULL,
	created_at TIMESTAMPTZ default now(),
	
	CONSTRAINT fk_document_a_id 
		FOREIGN KEY (doc_a_id) 
		REFERENCES documents(id) 
		ON DELETE CASCADE,

	CONSTRAINT fk_document_b_id 
		FOREIGN KEY (doc_b_id) 
		REFERENCES documents(id) 
		ON DELETE CASCADE
);
`
