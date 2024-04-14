# Virtualpaper

Virtualpaper is a document archiving solution that is heavily optimized for searching documents.
The biggest difference with Virtualpaper to many other solutions is that
Virtualpaper does not store the documents in folders. 
In fact there's no such entity as a folder in Virtualpaper.
How are documents located and filtered then?
Virtualpaper features user-configurable key-value metadata along with a very powerful and 
fast full-text-search to achieve the same effect, and much more.

For more information see the [official documentation](https://virtualpaper.tryffel.net).

The screenshot below showcases the most important aspect of Virtualpaper:
finding the documents you're looking for by typing any keywords, metadata or time ranges:
The interactive search suggests you with keywords as you type. 

![Screenshot](screenshot-document-search.png)


Rather than storing documents in a traditional folder structure,
the documents are simply stored in a single directory. 
The idea is to use metadata for storing the same relational information that the folder structure would encapsulate.
Instead of putting related documents to same folder or subfolder, 
Virtualpaper uses metadata key-values to indicate that the documents are somehow related.

For instance, instead of using folder structures like year and month, category, alphabets, 
all of this data can be stored in each document's metadata.
While this seems complicated and unintuitive, the benefit is clear:
instead of storing the documents in a single folder structure,
the documents now exist in several parallel contexts, just like folders.
Now documents can be filtered and sorted with any metadata or dates or their combinations.
Instead of navigating to the document by the folder structure like "it was probably under year 2022 and under invoices"
we can just query it with "date:2022 type:invoice", 
which will result in the same documents being listed.
Examples for multiple contexts are:
* List all 'invoices' from last year
* List all inquiries from company x that has value completed:false that are dated to time range
* List all documents related to a project
   
If you wish to benefit from this kind of filtering, you need to assign at least a few of these meaningful 
metadata-values. To help automate this, 
Virtualpaper tries to automatically match these values from document content when indexing them. 
In addition to filtering content according by metadata, Virtualpaper features full-text-search powered by Meilisearch,
which covers all metadata as well as content of the document itself.

This project is in **beta phase** and help with testing and general feedback is much appreciated.

## Features
* Store text documents (pdf, image files are extracted for text content)
* Save any use-configurable key-value metadata to documents
    * If configured, try to match key-values automatically from documents
    * Detect document date
    * User configurable rules for modifying the data
* REST api (swagger documentation is located at api/swaggerdocs/swagger.json) or at <virtualpaper-instance>/api/v1/swagger.json
* Full-text-search
* User-configurable rule engine for classifying documents and assigning metadata automatically either after creating or updating documents
* Responsive layout with dark theme
* **Total number of users is limited to 200.** This is because Meilisearch has a limit of 200 indices, and each user
uses one index. The benefit for own index is that each user can now configure their personal settings: 
  synonyms, stop words and results ranking, thus users have more powerful search capability over their files.
  Maybe one day it is possible to have more users, though.
* Option to add documents to favorites
* Share documents with individual users (read/write access)


## Requirements
Required 3rd party applications (run in docker, host, or another host machine):
* Postgresql
* Meilisearch v1.X

Create postgresql database and make sure to **initialize database as utf8** with e.g.: 
```CREATE DATABASE virtualpaper WITH ENCODING='utf8' TEMPLATE template0;```

Meilisearch does not require configuration other than from security perspective: consider setting apikey
and mode to production, and configure Virtualpaper accordingly. 
Meilisearch only indexes first 1000 words per document, which means that long documents
are not fully searchable by their content. 

# Building

## Server
You need Go 1.19 or later installed and configured.

Also for processing the documents you need Tesseract 5, Imagemagick 7, poppler-utils and optionally pandoc.
See Dockerfile for more info. 
Some distributions (e.g. Debian) ship Imagemagick-v6 by default. 
Please configure the locations for these executables in the configuration file. 

Build server with:
```make build```

## Frontend

Frontend is built with React and great React-Admin framework.
Make sure nodejs, npm and yarn are installed and then:

Initial configuration:
```cd frontend; npm install```

Build frontend with:
```make build-frontend```


# Configuration
Copy config.sample.toml to config.toml and place it to ~/.virtualpaper.toml.

Fill database and meilisearch configuration and you're good to go, at least for testing purposes.
All content is stored in filesystem, which is defined in config-file: Processing.data_dir.

All configuration variable can be overridden with environment variables, e.g.:
```VIRTUALPAPER_PROCESSING_DATA_DIR="/data"``` or
```VIRTUALPAPER_MEILISEARCH_URL="http://meilisearch:7700"```


# Run

See [documentation](https://virtualpaper.tryffel.net) for more help.


Virtualpaper can be run directly or with docker. 
Docker is easiest to get started with.

## Docker

The easiest way to get started is by using the provided docker-compose file:
```
docker-compose up
```

copy config.sample.toml to e.g. config-dir/config.toml.

By default, docker file includes only English-dataset for tesseract OCR engine. To use other languages,
either include them in Dockerfile, or install language packages on host machine and add them as volume to docker
with: ```-v /usr/share/tessdata:/usr/share/tessdata```. 
Host machine location may vary depending on distribution used.

Start server (for testing):
```docker run -d -v /config-dir:/config/ tryffel/virtualpaper:latest serve```

Start server (for persistence):
```
docker run -d \
    -v /config-dir:/config/ \
    -v /virtualpaper-data:/data \
    -v /virtualpaper-logs:/logs \
    tryffel/virtualpaper:latest serve
```

Create new user:
```
docker run -it \
    -v /config-dir:/config/ \
    -v /virtualpaper-data:/data \
    -v /usr/share/tessdata:/usr/share/tessdata \
    tryffel/virtualpaper:latest manage add-user
```

Reset password:
```
docker run -it \
    -v /config-dir:/config/ \
    -v /virtualpaper-data:/data \
    -v /usr/share/tessdata:/usr/share/tessdata \
    tryffel/virtualpaper:latest manage reset-password
```

## Manually
```virtualpaper --config config.toml serve```

# Usage

1. Create user with command 'manage add-user'.
2. Head over to web page, which is by default at http://localhost:8000 and login
3. Add some metadata key values. These are application-specific, but some initial keys might be
'correspondent', 'class', 'state', 'project' and fill some values for these. 
4. Upload documents on web page, let server index them and search for some documents.

# Development
See official docs for more info on how to get started.

Start frontend in development mode:
```make run-frontend```

Start backend:
```make run```

Spin up a development stack (this will start the server too, which can be stopped afterwards):
```make test-start```

Stop development stack:
```make test-stop```

## Tests (backend):

Unit tests:
```make test-unit```

Integration tests:
```make test-integration```

End-to-end tests (requires running server instance):
```make test-api```
e2e-tests communicate with the actual server and thus needs a working connection.
Before running e2e tests, start the server with ```make test-start```.
Also be sure the cleanup the server environment before running the e2e tests: ```make test-stop```.

All tests:
```make test```


## Develop backend with delve

First initialize the setup with ```make dev-init```.
Build the image ```make dev-build-container```.

A new directory dev/ is created. 
Only Virtualpaper-server is started. 
You will need to edit dev/config/config.toml 
to make sure Virtualpaper can connect to Postgresql and Meilisearch.

Launch the program with ```make dev-start-container```. 
Now delve is running and waiting for connection. 
Connect to delve from your IDE.

## License

This software is licensed under AGPL-v3.

