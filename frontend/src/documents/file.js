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

import * as React from "react";
import PropTypes from "prop-types";
import {useState} from "react";
import get from 'lodash/get';



export function downloadFile(url) {
    const  token  = localStorage.getItem('auth');
    return fetch(url, {
        method: "GET",
        headers: {"Authorization": `Bearer ${token}`}
    })
}

export function ThumbnailField ({ source, label, record }) {
    const url = get(record, source);
    const altText = get(record, label) ? get(record, label): "Thumbnail";

    const [imgData, setImage] = useState(() => {
        downloadFile(url)
            .then(response => {
                response.arrayBuffer().then(function (buffer) {
                    const data = window.URL.createObjectURL(new Blob([buffer]));
                    setImage(data);
                });
            })
            .catch(response => {
                    console.log(response);
                }
            );
        return "";
    });

    return (
        <div>
            <img src={imgData} alt={altText}/>
        </div>
    );
}

ThumbnailField.propTypes = {
    label: PropTypes.string,
    record: PropTypes.object,
    source: PropTypes.string.isRequired,
};


export function ThumbnailSmall ({ url, label })
{
    const [imgData, setImage] = useState(() => {
        downloadFile(url)
            .then(response => {
                response.arrayBuffer().then(function (buffer) {
                    const data = window.URL.createObjectURL(new Blob([buffer]));
                    setImage(data);
                });
            })
            .catch(response => {
                    console.log(response);
                }
            );
        return "";
    });

    return (
        <div style={{'overflow': 'hidden', 'max-height': '200px'}}>
            <img src={imgData} style={{'max-width': '250px'}} alt={label}/>
        </div>
    );
}


ThumbnailField.propTypes = {
    label: PropTypes.string,
    record: PropTypes.object,
    source: PropTypes.string.isRequired,
};

export function EmbedFile({ source, record })
{
    const url = get(record, source);
    console.log("Download file: ", url);
    const [imgData, setImage] = useState(() => {
        downloadFile(url)
            .then(response => {
                response.arrayBuffer().then(function (buffer) {
                    const data = window.URL.createObjectURL(new Blob([buffer]));
                    setImage(data);
                });
            })
            .catch( response => {
                    console.log(response);
                }
            );
        return "";
    });

    return (
        <div style={{display:'block', width:'100%'}}>
            <iframe style={{width: '100%', display: 'fill', border: 'none', height:'40em'}} title="Preview" src={imgData}/>
        </div>
    );
}

EmbedFile.propTypes = {
    label: PropTypes.string,
    record: PropTypes.object,
    source: PropTypes.string.isRequired,
};
