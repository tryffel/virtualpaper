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

const schemaV13 = `

CREATE TABLE auth_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    key TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,
    
    CONSTRAINT fk_user_id
		FOREIGN KEY (user_id) 
		REFERENCES users(id) 
		ON DELETE CASCADE,
		
	CONSTRAINT unique_token_key UNIQUE (key)
)`
