package tunnel

import (
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/khulnasoft/harbor-scanner-tunnel/pkg/etc"
	"github.com/khulnasoft/harbor-scanner-tunnel/pkg/ext"
	"github.com/stretchr/testify/require"
)

var (
	expectedReportJSON = `{
  "SchemaVersion": 2,
  "Results": [
    {
      "Target": "alpine:3.10.2",
      "Vulnerabilities": [
        {
          "VulnerabilityID": "CVE-2018-6543",
          "PkgName": "binutils",
          "InstalledVersion": "2.30-r1",
          "FixedVersion": "2.30-r2",
          "CVSS": {
            "nvd": {
              "V2Vector": "AV:L/AC:M/Au:N/C:P/I:N/A:N",
              "V3Vector": "CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:U/C:H/I:N/A:N",
              "V2Score": 1.9,
              "V3Score": 4.7
            },
            "redhat": {
              "V3Vector": "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
              "V3Score": 5.5
            }
          },
          "Severity": "MEDIUM",
          "References": [
            "https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-6543"
          ],
          "Layer": {
            "Digest": "sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"
          }
        }
      ]
    }
  ]
}`
	expectedReport = []Vulnerability{
		{
			VulnerabilityID:  "CVE-2018-6543",
			PkgName:          "binutils",
			InstalledVersion: "2.30-r1",
			FixedVersion:     "2.30-r2",
			Severity:         "MEDIUM",
			References: []string{
				"https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-6543",
			},
			Layer: &Layer{Digest: "sha256:5216338b40a7b96416b8b9858974bbe4acc3096ee60acbc4dfb1ee02aecceb10"},
			CVSS: map[string]CVSSInfo{
				"nvd": {
					V2Vector: "AV:L/AC:M/Au:N/C:P/I:N/A:N",
					V3Vector: "CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:U/C:H/I:N/A:N",
					V2Score:  float32Ptr(1.9),
					V3Score:  float32Ptr(4.7),
				},
				"redhat": {
					V2Vector: "",
					V3Vector: "CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
					V2Score:  nil,
					V3Score:  float32Ptr(5.5),
				},
			},
		},
	}

	expectedVersion = VersionInfo{
		Version: "v0.5.2-17-g3c9af62",
		VulnerabilityDB: &Metadata{
			NextUpdate: time.Unix(1584507644, 0).UTC(),
			UpdatedAt:  time.Unix(1584517644, 0).UTC(),
		},
	}
)

func TestWrapper_Scan(t *testing.T) {
	ambassador := ext.NewMockAmbassador()
	ambassador.On("Environ").Return([]string{"HTTP_PROXY=http://someproxy:7777"})
	ambassador.On("LookPath", "tunnel").Return("/usr/local/bin/tunnel", nil)

	config := etc.Tunnel{
		CacheDir:         "/home/scanner/.cache/tunnel",
		ReportsDir:       "/home/scanner/.cache/reports",
		DebugMode:        true,
		VulnType:         "os,library",
		SecurityChecks:   "vuln",
		Severity:         "CRITICAL,MEDIUM",
		IgnoreUnfixed:    true,
		IgnorePolicy:     "/home/scanner/opa/policy.rego",
		SkipUpdate:       true,
		SkipJavaDBUpdate: true,
		GitHubToken:      "<github_token>",
		Insecure:         true,
		Timeout:          5 * time.Minute,
	}

	imageRef := ImageRef{
		Name:     "alpine:3.10.2",
		Auth:     BasicAuth{Username: "dave.loper", Password: "s3cret"},
		Insecure: true,
	}

	expectedCmdArgs := []string{
		"/usr/local/bin/tunnel",
		"--cache-dir",
		"/home/scanner/.cache/tunnel",
		"--debug",
		"image",
		"--ignore-policy",
		"/home/scanner/opa/policy.rego",
		"--skip-java-db-update",
		"--skip-db-update",
		"--ignore-unfixed",
		"--no-progress",
		"--severity",
		"CRITICAL,MEDIUM",
		"--vuln-type",
		"os,library",
		"--scanners",
		"vuln",
		"--format",
		"json",
		"--output",
		"/home/scanner/.cache/reports/scan_report_1234567890.json",
		"alpine:3.10.2",
	}

	expectedCmdEnvs := []string{
		"HTTP_PROXY=http://someproxy:7777",
		"TUNNEL_TIMEOUT=5m0s",
		"TUNNEL_USERNAME=dave.loper",
		"TUNNEL_PASSWORD=s3cret",
		"TUNNEL_NON_SSL=true",
		"GITHUB_TOKEN=<github_token>",
		"TUNNEL_INSECURE=true",
	}

	ambassador.On("TempFile", "/home/scanner/.cache/reports", "scan_report_*.json").
		Return(ext.NewFakeFile("/home/scanner/.cache/reports/scan_report_1234567890.json", expectedReportJSON), nil)
	ambassador.On("Remove", "/home/scanner/.cache/reports/scan_report_1234567890.json").
		Return(nil)
	ambassador.On("RunCmd", &exec.Cmd{
		Path: "/usr/local/bin/tunnel",
		Env:  expectedCmdEnvs,
		Args: expectedCmdArgs},
	).Return([]byte{}, nil)

	report, err := NewWrapper(config, ambassador).Scan(imageRef)

	require.NoError(t, err)
	require.Equal(t, expectedReport, report)

	ambassador.AssertExpectations(t)
}

func TestWrapper_GetVersion(t *testing.T) {
	ambassador := ext.NewMockAmbassador()
	ambassador.On("LookPath", "tunnel").Return("/usr/local/bin/tunnel", nil)

	config := etc.Tunnel{
		CacheDir:  "/home/scanner/.cache/tunnel",
		DebugMode: true,
	}

	expectedCmdArgs := []string{
		"/usr/local/bin/tunnel",
		"--cache-dir",
		"/home/scanner/.cache/tunnel",
		"version",
		"--format",
		"json",
	}

	b, _ := json.Marshal(expectedVersion)
	ambassador.On("RunCmd", &exec.Cmd{
		Path: "/usr/local/bin/tunnel",
		Args: expectedCmdArgs},
	).Return(b, nil)

	vi, err := NewWrapper(config, ambassador).GetVersion()
	require.NoError(t, err)
	require.Equal(t, expectedVersion, vi)

	ambassador.AssertExpectations(t)
}

func float32Ptr(f float32) *float32 {
	return &f
}
