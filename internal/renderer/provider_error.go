package renderer

import (
	"github.com/riipandi/elph/internal/uiconst"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/riipandi/elph/pkg/core/agent"
)

func (m Model) addProviderErrorDetail(err error) Model {
	if err == nil {
		return m
	}
	body := provider.FormatProviderErrorDetail(err)
	body = agent.TruncateWithNotice(body, agent.MaxDisplayToolBytes)
	m = m.addDetailMessageWithStatus("Provider error", body, uiconst.DetailStatusError)
	m.session.AppendLog("provider_error", body)
	return m
}
