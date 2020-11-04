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
	tag_id INT,

	CONSTRAINT document_tags_pkey PRIMARY KEY (document_id, tag_id)
);


CREATE TABLE metadata_keys (
	id SERIAL,
	user_id INT,
	key TEXT NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT pk_metadata_keys PRIMARY KEY(id),
	CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
	CONSTRAINT unique_user_key UNIQUE (user_id, key)
);


CREATE TABLE metadata_values (
	id SERIAL,
	user_id INT,
	key_id INT,
	value TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT pk_metadata_values PRIMARY KEY(id),
	CONSTRAINT pk_metadata_values_key FOREIGN KEY (key_id) REFERENCES metadata_keys(id),
	CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
	CONSTRAINT unique_key_value UNIQUE (key_id, value)
);


CREATE TABLE document_metadata (
	document_id INT,
	key_id INT,
	value_id INT,

	CONSTRAINT pk_document_metaadata PRIMARY KEY(document_id, key_id, value_id),
	CONSTRAINT fk_document_metadata_doc_id FOREIGN KEY(document_id) REFERENCES documents(id),
	CONSTRAINT fk_keys FOREIGN KEY(key_id) REFERENCES metadata_keys(id),
	CONSTRAINT fk_values FOREIGN KEY(value_id) REFERENCES metadata_values(id)
);


`
