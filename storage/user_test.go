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

package storage

import (
	cache2 "github.com/patrickmn/go-cache"
	"reflect"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func TestUserStore_userCache(t *testing.T) {
	cache := cache2.New(time.Millisecond*10, time.Millisecond*10)

	user1 := &models.User{
		Timestamp: models.Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Id:       10,
		Name:     "Test user",
		Password: "pw",
		Email:    "test@user.com",
		IsAdmin:  false,
		IsActive: true,
	}

	store := UserStore{
		db:    nil,
		cache: cache,
	}
	tests := []struct {
		name    string
		wait    time.Duration
		setUser *models.User
		want    *models.User
	}{
		{
			name:    "simple user cache",
			wait:    time.Duration(0),
			setUser: user1,
			want:    user1,
		},
		{
			name:    "expired user",
			wait:    time.Millisecond * 30,
			setUser: user1,
			want:    nil,
		},
		{
			name:    "nil user",
			wait:    time.Duration(0),
			setUser: nil,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Flush()
			store.setUserCache(tt.setUser)
			if tt.wait != time.Duration(0) {
				time.Sleep(tt.wait)
				cache.DeleteExpired()
			}

			var userId int
			if tt.setUser != nil {
				userId = tt.setUser.Id
			}
			got := store.getUserIdCache(userId)
			ok := reflect.DeepEqual(got, tt.want)
			if !ok {
				t.Errorf("getUserIdCache() = %v, want %v", got, tt.want)
			}
		})
	}
}
