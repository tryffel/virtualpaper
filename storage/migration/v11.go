/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2023  Tero Vierimaa
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

const schemaV11 = `
CREATE TABLE document_view_history (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    document_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
                                   
	CONSTRAINT fk_user_id
		FOREIGN KEY (user_id) 
		REFERENCES users(id) 
		ON DELETE CASCADE,
		
	CONSTRAINT fk_document_id
		FOREIGN KEY (document_id) 
		REFERENCES documents(id) 
		ON DELETE CASCADE
);
`
