package providers

import (
	"context"
	"fmt"
	"os"
	"sort"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type HerokuClient interface {
	ConfigVarInfoForApp(ctx context.Context, appIdentity string) (heroku.ConfigVarInfoForAppResult, error)
	ConfigVarUpdate(ctx context.Context, appIdentity string, o map[string]*string) (heroku.ConfigVarUpdateResult, error)
}
type Heroku struct {
	client HerokuClient
	logger logging.Logger
}

func NewHeroku(logger logging.Logger) (core.Provider, error) {
	heroku.DefaultTransport.BearerToken = os.Getenv("HEROKU_API_KEY")

	svc := heroku.NewService(heroku.DefaultClient)
	return &Heroku{client: svc, logger: logger}, nil
}

func (h *Heroku) Name() string {
	return "heroku"
}

func (h *Heroku) Put(p core.KeyPath, val string) error {
	k := p.EffectiveKey()
	h.logger.WithField("path", p.Path).Debug("put variable")
	_, err := h.client.ConfigVarUpdate(context.TODO(), p.Path, map[string]*string{k: &val})
	return err
}
func (h *Heroku) PutMapping(p core.KeyPath, m map[string]string) error {
	vars := map[string]*string{}
	for k := range m {
		v := m[k]
		vars[k] = &v
	}
	h.logger.WithField("path", p.Path).Debug("put multiple values")
	_, err := h.client.ConfigVarUpdate(context.TODO(), p.Path, vars)
	return err
}

func (h *Heroku) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
	}

	k := secret

	entries := []core.EnvEntry{}
	for k, v := range k {
		val := ""
		if v != nil {
			val = *v
		}
		entries = append(entries, p.FoundWithKey(k, val))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (h *Heroku) Get(p core.KeyPath) (*core.EnvEntry, error) { // nolint:dupl
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
	}

	data := secret
	k := data[p.Env]
	if p.Field != "" {
		h.logger.WithField("path", p.Path).Debug("`env` attribute not found in returned data. take `field` attribute")
		k = data[p.Field]
	}

	if k == nil {
		h.logger.WithField("path", p.Path).Debug("requested entry not found")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*k)
	return &ent, nil
}

func (h *Heroku) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", h.Name())
}

func (h *Heroku) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", h.Name())
}

func (h *Heroku) getSecret(kp core.KeyPath) (heroku.ConfigVarInfoForAppResult, error) {
	h.logger.WithField("path", kp.Path).Debug("get field")
	return h.client.ConfigVarInfoForApp(context.TODO(), kp.Path)
}
