package main

import (
	"errors"
	"sync"
)

type Children map[string]Thinger // keyed by id

type Parent Thinger

type Thinger interface {
	GetId() string
	GetName() string
	Tag() string
	Parent() Parent
	SetParent(Parent)
	Children() Children
	AddChild(Thinger) error
	DeleteChild(Thinger) error
}

type Thing struct {
	Id         string
	Name       string
	parent     Parent
	childrenMu sync.RWMutex
	children   Children
}

func NewThing(id, name string) *Thing {
	return &Thing{
		Id:       id,
		Name:     name,
		children: make(Children),
	}
}

func (t *Thing) GetId() string           { return t.Id }
func (t *Thing) GetName() string         { return t.Name }
func (t *Thing) Parent() Parent          { return t.parent }
func (t *Thing) SetParent(parent Parent) { t.parent = parent }
func (t *Thing) Children() Children      { return t.children }

func (t *Thing) Tag() string {
	if t.parent == nil {
		return t.Id
	}
	return t.parent.Tag() + "." + t.Id
}

func (t *Thing) AddChild(child Thinger) error {
	t.childrenMu.Lock()
	defer t.childrenMu.Unlock()
	if _, exists := t.children[child.GetId()]; exists {
		return errors.New("child already exists")
	}

	t.children[child.GetId()] = child
	child.SetParent(t)
	return nil
}

func (t *Thing) DeleteChild(child Thinger) error {
	t.childrenMu.Lock()
	defer t.childrenMu.Unlock()
	if _, exists := t.children[child.GetId()]; !exists {
		return errors.New("child does not exist")
	}
	delete(t.children, child.GetId())
	child.SetParent(nil)
	return nil
}
