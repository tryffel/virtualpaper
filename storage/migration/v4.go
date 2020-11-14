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

const schemaV4 = `
CREATE TABLE process_rules (
	id SERIAL PRIMARY KEY,
	user_id INT,
	rule_type TEXT NOT NULL,
	filter TEXT NOT NULL,
	comment TEXT NOT NULL default '',
	action TEXT NOT NULL default '',
	active BOOL NOT NULL DEFAULT true,

	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
	CONSTRAINT rules_user_filter_unique UNIQUE (user_id, filter)
);
`

// This needs to be run manually and uuids need to be generated with external tool and inserted after first statement.
// This only applies if there is existing schema with v3 before commit bf4e638c .
const schemaDocumentIdToUuid = `
alter table documents add column uuid TEXT;

--- set document uuids by hand, generate with online tool if necessary.

-- generate new column
alter table jobs add column document_uuid TEXT;
alter table process_queue add column document_uuid TEXT;
alter table document_tags add column document_uuid TEXT;
alter table document_metadata add column document_uuid TEXT;


-- populate data
update jobs
set document_uuid=documents.uuid
from documents
where documents.id=jobs.document_id;

update document_tags t
set document_uuid=documents.uuid
from documents
where documents.id=t.document_id;

update document_metadata t
set document_uuid=documents.uuid
from documents
where documents.id=t.document_id;


-- drop old constraints
alter table jobs drop constraint fk_document;
alter table process_queue drop constraint fk_document;
alter table document_tags drop constraint document_tags_pkey;
alter table document_metadata drop constraint fk_document_metadata_doc_id;

--- rename uuid to id
alter table jobs drop column document_id; alter table jobs rename document_uuid to document_id;
alter table process_queue drop column document_id; alter table process_queue rename document_uuid to document_id;
alter table document_tags drop column document_id; alter table document_tags rename document_uuid to document_id;
alter table document_metadata drop column document_id; alter table document_metadata rename document_uuid to document_id;

alter table documents drop constraint documents_pkey cascade;
alter table documents drop column id;
alter table documents
rename column uuid to id;

-- add constraints
alter table
documents add constraint document_pkey PRIMARY KEY (id);

alter table jobs add CONSTRAINT fk_document FOREIGN KEY(document_id)
REFERENCES documents(id);

alter table process_queue add CONSTRAINT pk_queue PRIMARY KEY(document_id, step);
alter table process_queue add CONSTRAINT fk_document FOREIGN KEY(document_id) REFERENCES documents(id);

alter table document_tags
add CONSTRAINT document_tags_pkey PRIMARY KEY (document_id, tag_id);

alter table document_metadata
add CONSTRAINT pk_document_metadata PRIMARY KEY(document_id, key_id, value_id);

alter table document_metadata
add 	CONSTRAINT fk_document_metadata_doc_id FOREIGN KEY(document_id) REFERENCES documents(id);
`
