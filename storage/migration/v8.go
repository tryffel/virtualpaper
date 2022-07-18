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

const schemaV8 = `
ALTER TABLE metadata_keys
DROP CONSTRAINT fk_user,
ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) 
	REFERENCES users(id)
	ON DELETE CASCADE;
	

ALTER TABLE metadata_values
DROP CONSTRAINT fk_user,
ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) 
	REFERENCES users(id)
	ON DELETE CASCADE;
	
	
ALTER TABLE process_rules
DROP CONSTRAINT fk_user,
ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) 
	REFERENCES users(id)
	ON DELETE CASCADE;
	

ALTER TABLE rules
DROP CONSTRAINT fk_user,
ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) 
	REFERENCES users(id)
	ON DELETE CASCADE;
	

ALTER TABLE tags
DROP CONSTRAINT fk_user,
ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) 
	REFERENCES users(id)
	ON DELETE CASCADE;
	

ALTER TABLE user_preferences
DROP CONSTRAINT fk_user,
ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) 
	REFERENCES users(id)
	ON DELETE CASCADE;
	

ALTER TABLE document_metadata
-- document_metadata previously had a duplicate 
-- constraint fk_keys and fk_metadata_values_key
DROP CONSTRAINT fk_values,
ADD CONSTRAINT fk_values 
	FOREIGN KEY (value_id) 
	REFERENCES metadata_values(id)
	ON DELETE CASCADE,
DROP CONSTRAINT fk_metadata_values_key,
DROP CONSTRAINT fk_keys,
ADD CONSTRAINT fk_keys 
	FOREIGN KEY (key_id)
	REFERENCES metadata_keys(id)
	ON DELETE CASCADE;
	
ALTER TABLE rule_actions
DROP CONSTRAINT fk_metadata_value,
ADD CONSTRAINT fk_metadata_value 
	FOREIGN KEY (metadata_value) 
	REFERENCES metadata_values(id)
	ON DELETE CASCADE,
	
DROP CONSTRAINT fk_metadata_key,
ADD CONSTRAINT fk_metadata_key 
	FOREIGN KEY (metadata_key) 
	REFERENCES metadata_keys(id)
	ON DELETE CASCADE;
	
	
ALTER TABLE rule_conditions
DROP CONSTRAINT fk_metadata_key,
ADD CONSTRAINT fk_metadata_key 
	FOREIGN KEY (metadata_key) 
	REFERENCES metadata_keys(id)
	ON DELETE CASCADE,
	
DROP CONSTRAINT fk_metadata_value,
ADD CONSTRAINT fk_metadata_value 
	FOREIGN KEY (metadata_value) 
	REFERENCES metadata_values(id)
	ON DELETE CASCADE;
`
