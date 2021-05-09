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

import React, { Component } from 'react';
import { connect } from 'react-redux';
import { crudGetOne, UserMenu, MenuItemLink } from 'react-admin';
import SettingsIcon from '@material-ui/icons/Settings';
import AccountCircle from '@material-ui/icons/AccountCircle';
import authProvider from "./authProvider";

class MyUserMenuView extends Component {
    componentDidMount() {
        this.fetchProfile();
    }

    fetchProfile = () => {
        authProvider.checkAuth().then((x) => {
            this.props.crudGetOne(
                'preferences',
                'user',
                '/preferences',
                true
            )}).catch((error) => {
            console.error(error)
        });
    };

    render() {
        const { crudGetOne, profile, admin, ...props } = this.props;
        return (
            <UserMenu label={profile ? profile.user_name: ''} {...props}>
                <MenuItemLink
                    to="/preferences"
                    primaryText="Preferences"
                    leftIcon={<AccountCircle/>}
                />

                {authProvider.isAdmin() ?
                    <MenuItemLink
                        to="/admin"
                        primaryText="Administrator"
                        leftIcon={<SettingsIcon/>}
                    />: null
                }
            </UserMenu>
        );
    }
}

const mapStateToProps = state => {
    const resource = 'preferences';
    const id = 'user';

    if (state.admin.resources[resource]) {
        if (state.admin.resources[resource].data[id]) {
            localStorage.setItem("is_admin", state.admin.resources[resource].data[id].is_admin);
        }
    }

    return {
        profile: state.admin.resources[resource]
            ? state.admin.resources[resource].data[id]
            : null
    };
};

const MyUserMenu = connect(
    mapStateToProps,
    { crudGetOne }
)(MyUserMenuView);
export default MyUserMenu;
