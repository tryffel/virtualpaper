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

const schemaV3 = `
CREATE TABLE metadata (
	document_id INT,
	
	key TEXT NOT NULL,
	key_lower TEXT NOT NULL,

	value TEXT NOT NULL,
	value_lower TEXT NOT NULL,

	CONSTRAINT pk_metadata PRIMARY KEY (document_id, key_lower),
	CONSTRAINT fk_document FOREIGN KEY(document_id) REFERENCES documents(id)
);


CREATE TABLE tags (
	id SERIAL,
	user_id INT,
	key TEXT NOT NULL,
	comment TEXT NOT NULL DEFAULT '',
  	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
	CONSTRAINT tags_user_key_unique UNIQUE (user_id, key),
	CONSTRAINT tags_pkey PRIMARY KEY (id)
);


CREATE TABLE document_tags (
	document_id INT,
	tag_id,

	CONSTRAINT document_tags_pkey PRIMARY KEY (document_id, tag_id)
);
`
