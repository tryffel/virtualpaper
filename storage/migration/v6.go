/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2021  Tero Vierimaa
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

const schemaV6 = `
-- jobs
ALTER TABLE jobs DROP CONSTRAINT fk_document;
ALTER TABLE jobs ADD CONSTRAINT fk_document FOREIGN KEY(document_id)
	REFERENCES documents(id) ON DELETE CASCADE;

-- metadata
ALTER TABLE document_metadata DROP CONSTRAINT fk_document_metadata_doc_id;

ALTER TABLE document_metadata ADD CONSTRAINT fk_document_metadata_doc_id
	FOREIGN KEY(document_id) REFERENCES documents(id) ON DELETE CASCADE;

ALTER TABLE metadata_values DROP CONSTRAINT pk_metadata_values_key;
ALTER TABLE document_metadata ADD CONSTRAINT fk_metadata_values_key 
	FOREIGN KEY (key_id) REFERENCES metadata_keys(id) ON DELETE CASCADE;

-- process queue
ALTER TABLE process_queue DROP CONSTRAINT fk_document;
ALTER TABLE process_queue ADD CONSTRAINT fk_document FOREIGN KEY(document_id) 
	REFERENCES documents(id) ON DELETE CASCADE;

-- documents
ALTER TABLE documents ADD COLUMN deleted_at TIMESTAMPTZ;
`
