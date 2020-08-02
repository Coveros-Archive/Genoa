package v3

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"testing"
)

func Test_getIdxOfChartVersionFromChartEntries(t *testing.T) {
	type args struct {
		slice         repo.ChartVersions
		lookupVersion string
		left          int
		right         int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "chart version exists at idx 6",
			args: args{
				slice: repo.ChartVersions{
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
				},
				lookupVersion: "2.0.2",
				left:          0,
				right:         6,
			},
			want:    6,
			wantErr: false,
		},
		{
			name: "chart version exists at idx 0",
			args: args{
				slice: repo.ChartVersions{
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
				},
				lookupVersion: "1.0.1",
				left:          0,
				right:         6,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "chart version exists at idx 2",
			args: args{
				slice: repo.ChartVersions{
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
				},
				lookupVersion: "1.1.0",
				left:          0,
				right:         6,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "chart version exists at idx 3",
			args: args{
				slice: repo.ChartVersions{
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
				},
				lookupVersion: "1.1.1",
				left:          0,
				right:         6,
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "chart version exists at idx 4",
			args: args{
				slice: repo.ChartVersions{
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
				},
				lookupVersion: "2.0.0",
				left:          0,
				right:         6,
			},
			want:    4,
			wantErr: false,
		},
		{
			name: "chart version does not",
			args: args{
				slice: repo.ChartVersions{
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "1.1.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.0",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.1",
						},
						URLs: []string{"https://download-url.com"},
					},
					&repo.ChartVersion{
						Metadata: &chart.Metadata{
							Name:    "jenkins",
							Version: "2.0.2",
						},
						URLs: []string{"https://download-url.com"},
					},
				},
				lookupVersion: "100.1.01",
				left:          0,
				right:         6,
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIdxOfChartVersionFromChartEntries(tt.args.slice, tt.args.lookupVersion, tt.args.left, tt.args.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIdxOfChartVersionFromChartEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("getIdxOfChartVersionFromChartEntries() got = %v, want %v", got, tt.want)
			}
		})
	}
}
