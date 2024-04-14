# Changelog

## Unreleased

### Added
- Initial support for sharing documents between users. User can now share read/write access to other users.
- Add support for document favorites to quickly pin documents
- Implement running rules after document has been updated. Each rule can now be configure to run either after document has been created, updated or both.
- Implement filtering rule list
- Implement searching metadata key-values

### Changed
- All api endpoints now return http 404 instead of 403 when user is unauthorized
- Implement service layer inside all api calls
- Implement using database transactions throughout operations
- Bumpded Meilisearch version to v.1.6
- Frontend is now bundled using `vite` & `esbuild`
- Lots of improvements in frontend layouts and dark theme

### Fixed:
- fixed broken CLI command `index`


## Release 0.5 - 2023-11-29

### Features
- Add support for detecting document languages. This is run automatically when uploading new documents.
- Added dark theme
- Added icon and option for setting color to metadata keys

### Improvements
- Added tooltip to show documents when hovering over metadata keys in document page
- Improved the UI layouts for several pages: document show, user preferences, admin page
- Improvements to background processing schema
- Improved logging: added correlation ids (requestId, taskId)
- Removed deprecated Inotify-based file watch


**Full Changelog**: https://github.com/tryffel/virtualpaper/compare/v0.4.0...v0.5.0


### Docker image:

- `tryffel/virtualpaper:0.5`
- `tryffel/virtualpaper:0.5-arm64`


## Release 0.4 - 2023-05-14
---

Docker image:
- `tryffel/virtualpaper:0.4`
- `tryffel/virtualpaper:0.4-arm64`.

### Features
- Add modal for reordering processing rules

### Improvements
- Documents are unique but scoped to each user, allowing different users to still upload same document
- Added link 'forgot password' on login page
- Added 'upload documents' button to dashboard page
- Documents's size is now shown on document page
- Improved query parser:
    - Phrases are now detected correctly
    - Adding metadata filters ANDs them together unless the query contains another operator already


## Release 0.3 - 2023-04-15
---


### Features
- auth tokens are persisted and can be revoked
- administrator can create new users
- administrator can edit existing users
- added download-button to download the original document
- added document trash bin: deleting documents now result in documents going to separate trash bin, where user can restore them. Document will be automatically deleted from the system after certain time (default: 14 days)

### Improvements
- only active users can login
- login is case-insensitive
- Hide most actions behind menu when viewing document
- Show document's size
- Improve document view layout especially with small displays


### Other
- enforce usernames: must be alphanumeric, whitespaces are allowed
- enforce emails: must be unique


### Updating

#### Before updating

See [instructions for v0.3](/more/v0.3-checks/#updating-to-v0.3).

1. **Backup the database**
1. verify that usernames are unique do not clash and fix manually if necessary
1. verify that emails are unique and do not clash and fix manually if necessary
1. delete the `api.secret_key` from config file

Use docker image `tryffel/virtualpaper:v0.3` or `tryffel/virtualpaper:v0.3-arm64`.


**Note**: this update will introduce database migrations that may potentially fail.
To make sure the migration goes smoothly, please **run following checks before doing the update**.
Only proceed updating if the checks pass, and test again after you have fixed possible conflicts.
If there are only a few users, it's okay to check the values from the admin UI, but the database query
will give exact results.


## Release 0.2 - 2023-03-22
---

Use docker image `tryffel/virtualpaper:v0.2` or `tryffel/virtualpaper:v0.2-arm64`.


### Features
- New search bar with suggestions
- Allow user to reset their password with email link
- Link documents
- Dashboard now shows last viewed documents in addition to last created and last updated documents
- Document history, which saves edits for documents and their metadata

### Improvements
- Improve search bar, parse utf8 characters correctly
- Handle whitespaces in metadata keys and values
- Tesseract and Imagemagick are now called with `exec` instead of statically linking them. This makes a lot easier to deploy and to develop Virtualpaper.
- Improved validation for API
- Improved the document list and show layouts. Mimetypes and dates are now visible as chips.
- Indexing status is now shown on document list toolbar, if indexing is ongoing.
- Improved rule edit view
- Metadata key list now has document count included in addition to metadata value count
- Improved 'schedule document' feature both for admins and for users
- Added rate limit to all authentication endpoints

### Other
- New api test suite
- Basic CI pipeline completed for running tests
- Rewrote the API layer with Echo router
- Meilisearch is now on version 1


## Release 0.1 - 2022-07-19
---

Initial release.

Use docker image `tryffel/virtualpaper:v0.1`.
