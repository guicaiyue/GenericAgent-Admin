package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type channelField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Secret      bool   `json:"secret,omitempty"`
	Type        string `json:"type,omitempty"`
	Value       string `json:"value,omitempty"`
	HasValue    bool   `json:"has_value,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

type channelProfile struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Fields      []channelField `json:"fields"`
}

type channelsResponse struct {
	Path     string           `json:"path"`
	Exists   bool             `json:"exists"`
	Profiles []channelProfile `json:"profiles"`
}

var channelDefinitions = []channelProfile{
	{ID: "feishu", Name: "飞书 / Lark", Description: "机器人应用凭据、用户白名单与公开访问开关。", Fields: []channelField{
		{Name: "fs_app_id", Label: "App ID", Placeholder: "cli_xxx"},
		{Name: "fs_app_secret", Label: "App Secret", Secret: true, Placeholder: "留空则保留 mykey.py 已有值"},
		{Name: "fs_allowed_users", Label: "Allowed Users", Placeholder: "user1,user2 或留空", Type: "list"},
		{Name: "fs_public_access", Label: "Public Access", Type: "bool"},
	}},
	{ID: "wecom", Name: "企业微信", Description: "企业微信机器人/应用通道凭据与允许用户。", Fields: []channelField{
		{Name: "wecom_bot_id", Label: "Bot ID / Agent ID", Placeholder: "ww/openapi id"},
		{Name: "wecom_secret", Label: "Secret", Secret: true, Placeholder: "留空则保留 mykey.py 已有值"},
		{Name: "wecom_allowed_users", Label: "Allowed Users", Placeholder: "user1,user2 或留空", Type: "list"},
	}},
	{ID: "dingtalk", Name: "钉钉", Description: "钉钉机器人/应用通道 client 凭据与允许用户。", Fields: []channelField{
		{Name: "dingtalk_client_id", Label: "Client ID", Placeholder: "dingxxx"},
		{Name: "dingtalk_client_secret", Label: "Client Secret", Secret: true, Placeholder: "留空则保留 mykey.py 已有值"},
		{Name: "dingtalk_allowed_users", Label: "Allowed Users", Placeholder: "user1,user2 或留空", Type: "list"},
	}},
}

func (s *Server) channels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.loadChannelsResponse())
	case http.MethodPut:
		if r.Header.Get("X-GA-Confirm") != "dangerous" {
			bad(w, http.StatusPreconditionRequired, "dangerous operation requires X-GA-Confirm: dangerous")
			return
		}
		var req struct {
			Profiles []channelProfile `json:"profiles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			bad(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.saveChannels(req.Profiles); err != nil {
			bad(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, s.loadChannelsResponse())
	default:
		bad(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) channelConfigPath() string {
	return filepath.Join(s.CfgStore.Cfg.GARoot, "mykey.py")
}

func (s *Server) loadChannelsResponse() channelsResponse {
	values := map[string]string{}
	exists := false
	if strings.TrimSpace(s.CfgStore.Cfg.GARoot) != "" {
		if content, err := os.ReadFile(s.channelConfigPath()); err == nil {
			exists = true
			values = parseChannelAssignments(string(content))
		}
	}
	profiles := cloneChannelDefinitions()
	for pi := range profiles {
		for fi := range profiles[pi].Fields {
			f := &profiles[pi].Fields[fi]
			if v, ok := values[f.Name]; ok {
				f.HasValue = strings.TrimSpace(v) != "" && v != "[]" && strings.ToLower(v) != "false"
				if !f.Secret {
					f.Value = normalizeChannelDisplayValue(v, f.Type)
				}
			}
		}
	}
	return channelsResponse{Path: s.channelConfigPath(), Exists: exists, Profiles: profiles}
}

func (s *Server) saveChannels(profiles []channelProfile) error {
	if strings.TrimSpace(s.CfgStore.Cfg.GARoot) == "" {
		return fmt.Errorf("GA root is not configured")
	}
	path := s.channelConfigPath()
	content := ""
	existing := map[string]string{}
	if b, err := os.ReadFile(path); err == nil {
		content = string(b)
		existing = parseChannelAssignments(content)
	} else if !os.IsNotExist(err) {
		return err
	}
	incoming := map[string]channelField{}
	for _, p := range profiles {
		for _, f := range p.Fields {
			incoming[f.Name] = f
		}
	}
	values := map[string]string{}
	for _, p := range channelDefinitions {
		for _, def := range p.Fields {
			f, ok := incoming[def.Name]
			if !ok {
				f = def
			}
			if def.Secret && f.Value == "" {
				values[def.Name] = existing[def.Name]
			} else {
				values[def.Name] = encodeChannelValue(f.Value, def.Type)
			}
		}
	}
	updated := upsertChannelAssignments(content, values)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(updated), 0600)
}

func cloneChannelDefinitions() []channelProfile {
	out := make([]channelProfile, len(channelDefinitions))
	for i, p := range channelDefinitions {
		out[i] = p
		out[i].Fields = append([]channelField{}, p.Fields...)
	}
	return out
}

var assignRe = regexp.MustCompile(`(?m)^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.+?)\s*$`)

var pyStrTokenRe = regexp.MustCompile(`'((?:[^'\\]|\\.)*)'|"((?:[^"\\]|\\.)*)"`)

// stripPyComment removes a trailing Python `#` comment that lies outside of any
// string literal. It returns the code portion of the line untouched.
func stripPyComment(s string) string {
	inStr := false
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			if c == '\\' {
				i++
				continue
			}
			if c == quote {
				inStr = false
			}
			continue
		}
		switch c {
		case '\'', '"':
			inStr = true
			quote = c
		case '#':
			return s[:i]
		}
	}
	return s
}

// bracketDelta returns the net change in open-bracket depth for s, ignoring any
// brackets that appear inside string literals.
func bracketDelta(s string) int {
	depth := 0
	inStr := false
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			if c == '\\' {
				i++
				continue
			}
			if c == quote {
				inStr = false
			}
			continue
		}
		switch c {
		case '\'', '"':
			inStr = true
			quote = c
		case '[', '{', '(':
			depth++
		case ']', '}', ')':
			depth--
		}
	}
	return depth
}

func pyUnescape(s string) string {
	s = strings.ReplaceAll(s, `\\`, "\x00")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\'`, `'`)
	s = strings.ReplaceAll(s, "\x00", `\`)
	return s
}

// collectAssignmentValue gathers the (possibly multi-line) right-hand side of an
// assignment whose first value fragment is first. It returns the comment-stripped
// joined value and the index of the last line consumed.
func collectAssignmentValue(lines []string, start int, first string) (string, int) {
	frag := strings.TrimSpace(stripPyComment(first))
	depth := bracketDelta(frag)
	parts := []string{frag}
	end := start
	for depth > 0 && end+1 < len(lines) {
		end++
		seg := strings.TrimSpace(stripPyComment(lines[end]))
		depth += bracketDelta(seg)
		if seg != "" {
			parts = append(parts, seg)
		}
	}
	return strings.Join(parts, " "), end
}

func parseChannelAssignments(content string) map[string]string {
	out := map[string]string{}
	allowed := map[string]string{}
	for _, p := range channelDefinitions {
		for _, f := range p.Fields {
			allowed[f.Name] = f.Type
		}
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		m := assignRe.FindStringSubmatch(lines[i])
		if len(m) != 3 {
			continue
		}
		typ, ok := allowed[m[1]]
		if !ok {
			continue
		}
		value, end := collectAssignmentValue(lines, i, m[2])
		out[m[1]] = normalizeChannelDisplayValue(value, typ)
		i = end
	}
	return out
}

func upsertChannelAssignments(content string, values map[string]string) string {
	allowed := map[string]bool{}
	formatted := map[string]string{}
	for _, p := range channelDefinitions {
		for _, f := range p.Fields {
			allowed[f.Name] = true
			formatted[f.Name] = fmt.Sprintf("%s = %s", f.Name, formatPythonLiteral(values[f.Name], f.Type))
		}
	}
	seen := map[string]bool{}
	raw := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if strings.TrimSpace(content) == "" {
		raw = []string{}
	}
	lines := []string{}
	for i := 0; i < len(raw); i++ {
		m := assignRe.FindStringSubmatch(raw[i])
		if len(m) == 3 && allowed[m[1]] {
			_, end := collectAssignmentValue(raw, i, m[2])
			lines = append(lines, formatted[m[1]])
			seen[m[1]] = true
			i = end
			continue
		}
		lines = append(lines, raw[i])
	}
	if strings.TrimSpace(content) != "" && (len(lines) == 0 || strings.TrimSpace(lines[len(lines)-1]) != "") {
		lines = append(lines, "")
	}
	missing := []string{}
	for _, p := range channelDefinitions {
		for _, f := range p.Fields {
			if !seen[f.Name] {
				missing = append(missing, formatted[f.Name])
			}
		}
	}
	if len(missing) > 0 {
		lines = append(lines, "# GA Admin channel configuration", "# Managed by GA Admin; secret values are masked in the UI.")
		lines = append(lines, missing...)
	}
	out := strings.Join(lines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	return out
}

func normalizeChannelDisplayValue(raw, typ string) string {
	raw = strings.TrimSpace(raw)
	switch typ {
	case "bool":
		return strings.ToLower(raw)
	case "list":
		inner := raw
		if i := strings.Index(raw, "["); i >= 0 {
			if j := strings.LastIndex(raw, "]"); j > i {
				inner = raw[i+1 : j]
			}
		}
		if matches := pyStrTokenRe.FindAllStringSubmatch(inner, -1); len(matches) > 0 {
			items := make([]string, 0, len(matches))
			for _, mm := range matches {
				tok := mm[1]
				if tok == "" && mm[2] != "" {
					tok = mm[2]
				}
				items = append(items, pyUnescape(tok))
			}
			return strings.Join(items, ",")
		}
		return strings.Trim(strings.TrimSpace(inner), "'\"")
	default:
		if v, err := strconv.Unquote(raw); err == nil {
			return v
		}
		if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
			return pyUnescape(raw[1 : len(raw)-1])
		}
		return strings.Trim(raw, "'\"")
	}
}

func encodeChannelValue(v, typ string) string {
	v = strings.TrimSpace(v)
	if typ == "bool" {
		if strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes") || strings.EqualFold(v, "on") {
			return "true"
		}
		return "false"
	}
	return v
}

func formatPythonLiteral(v, typ string) string {
	switch typ {
	case "bool":
		if strings.EqualFold(v, "true") {
			return "True"
		}
		return "False"
	case "list":
		parts := []string{}
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				parts = append(parts, part)
			}
		}
		sort.Strings(parts)
		b, _ := json.Marshal(parts)
		return string(b)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
