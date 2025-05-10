package main

import (
	"context"
)

var weaponsData = weaponCache{}

type weaponCache struct {
	data map[int]weaponCacheRow
}
type weaponCacheRow struct {
	Id          int
	Name        string
	VisibleName string
	Emote       string
}

func (c *weaponCache) GetWeapon(id int) *weaponCacheRow {
	if c.data == nil {
		c.LoadWeaponCache()
	}
	data, ok := c.data[id]

	if ok {
		return &data
	}
	return &weaponCacheRow{}
}
func (c *weaponCache) GetWeaponByName(name string) *weaponCacheRow {
	if c.data == nil {
		c.LoadWeaponCache()
	}
	for _, v := range c.data {
		if v.Name == name {
			return &v
		}
	}
	return &weaponCacheRow{}
}
func (c *weaponCache) GetAllWeapons() *[]weaponCacheRow {
	if c.data == nil {
		c.LoadWeaponCache()
	}
	cacheData := []weaponCacheRow{}
	for _, row := range c.data {
		cacheData = append(cacheData, row)
	}
	return &cacheData
}

func (c *weaponCache) LoadWeaponCache() error {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()
	weaponRows, err := conn.Query(context.Background(), "select id, name, visible_name, emote from weapons")
	if err != nil {
		log.Error("Could not retrieve weapons from database", "error", err)
		return err
	} else {
		c.data = map[int]weaponCacheRow{}
		for weaponRows.Next() {
			var id int
			var value string
			var label string
			var emote string
			err := weaponRows.Scan(&id, &value, &label, &emote)
			if err != nil {
				log.Error("Could not scan weapon row", "error", err)
				continue
			}
			c.data[id] = weaponCacheRow{Id: id, Name: value, VisibleName: label, Emote: emote}
		}
	}
	return nil
}

var rolesData = roleCache{}

type roleCache struct {
	data map[int]roleCacheRow
}
type roleCacheRow struct {
	Id          int
	Name        string
	VisibleName string
	Emote       string
}

func (c *roleCache) GetRole(id int) *roleCacheRow {
	if c.data == nil {
		c.LoadRoleCache()
	}
	data, ok := c.data[id]

	if ok {
		return &data
	}
	return &roleCacheRow{}
}
func (c *roleCache) GetRoleByName(name string) *roleCacheRow {
	if c.data == nil {
		c.LoadRoleCache()
	}
	for _, v := range c.data {
		if v.Name == name {
			return &v
		}
	}
	return &roleCacheRow{}
}
func (c *roleCache) GetAllRoles() *[]roleCacheRow {
	if c.data == nil {
		c.LoadRoleCache()
	}
	cacheData := []roleCacheRow{}
	for _, row := range c.data {
		cacheData = append(cacheData, row)
	}
	return &cacheData
}

func (c *roleCache) LoadRoleCache() error {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()
	rolesRows, err := conn.Query(context.Background(), "select id, name, visible_name, emote from roles")
	if err != nil {
		log.Error("Could not retrieve roles from database", "error", err)
		return err
	} else {
		c.data = map[int]roleCacheRow{}
		for rolesRows.Next() {
			var id int
			var value string
			var label string
			var emote string
			err := rolesRows.Scan(&id, &value, &label, &emote)
			if err != nil {
				log.Error("Could not scan weapon row", "error", err)
				continue
			}
			c.data[id] = roleCacheRow{Id: id, Name: value, VisibleName: label, Emote: emote}
		}
	}
	return nil
}
