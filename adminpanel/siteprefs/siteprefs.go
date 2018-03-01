package siteprefs

import (
	"fmt"
	"net/http"

	"github.com/gocraft/dbr"
)

// SitePref is the model object corresponding to the site_pref table.
type SitePref struct {
	PrefKey   string `db:"pref_key" json:"pref_key" yaml:"pref_key"`
	PrefValue string `db:"pref_value" json:"pref_value" yaml:"pref_value"`
}

type SitePrefsModel struct {
	DBDriverName string // TODO: autowire struct tag
	DBDSN        string // TODO: autowire struct tag
	TablePrefix  string
	Connection   *dbr.Connection
	// TODO: figure out how file sync stuff works
}

func (m *SitePrefsModel) AfterWire() error {

	if m.DBDriverName == "" {
		return fmt.Errorf("SitePrefsModel.DBDriverName cannot be empty")
	}
	if m.DBDSN == "" {
		return fmt.Errorf("SitePrefsModel.DBDSN cannot be empty")
	}

	// FIXME: hm, we could register our migrations here... so we can use the table prefix

	conn, err := dbr.Open(m.DBDriverName, m.DBDSN, nil)
	if err != nil {
		return err
	}
	m.Connection = conn
	return nil
}

func (m *SitePrefsModel) ReadAll() ([]SitePref, error) {
	ssn := m.Connection.NewSession(nil)
	ssn.Select("*").From("")
	return nil, nil
}

func (m *SitePrefsModel) Read(prefKey string) (*SitePref, error) {
	return nil, nil
}

func (m *SitePrefsModel) Write(prefKey, prefValue string) error {
	return nil
}

type SitePrefsHandler struct {
	Model *SitePrefsModel
}

func NewSitePrefsHandler(model *SitePrefsModel) *SitePrefsHandler {
	return &SitePrefsHandler{
		Model: model,
	}
}

func (h *SitePrefsHandler) ServeHTTPChain(w http.ResponseWriter, r *http.Request) (wret http.ResponseWriter, rret *http.Request) {

	wret = w
	rret = r

	return
}
