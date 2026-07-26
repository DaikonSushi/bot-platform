package pluginmgr

import "testing"

func TestParseGitHubRepo(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantURL   string
		wantErr   bool
	}{
		{
			name:      "https URL",
			input:     "https://github.com/DaikonSushi/plugin-agent",
			wantOwner: "DaikonSushi",
			wantRepo:  "plugin-agent",
			wantURL:   "https://github.com/DaikonSushi/plugin-agent",
		},
		{
			name:      "short owner repo",
			input:     "DaikonSushi/plugin-agent",
			wantOwner: "DaikonSushi",
			wantRepo:  "plugin-agent",
			wantURL:   "https://github.com/DaikonSushi/plugin-agent",
		},
		{
			name:      "git ssh URL",
			input:     "git@github.com:DaikonSushi/plugin-agent.git",
			wantOwner: "DaikonSushi",
			wantRepo:  "plugin-agent",
			wantURL:   "https://github.com/DaikonSushi/plugin-agent",
		},
		{
			name:    "invalid",
			input:   "https://example.com/DaikonSushi/plugin-agent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, normalized, err := parseGitHubRepo(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if owner != tt.wantOwner || repo != tt.wantRepo || normalized != tt.wantURL {
				t.Fatalf("parseGitHubRepo() = (%q, %q, %q), want (%q, %q, %q)", owner, repo, normalized, tt.wantOwner, tt.wantRepo, tt.wantURL)
			}
		})
	}
}
