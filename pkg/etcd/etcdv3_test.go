package etcd

import "testing"

func Test_getRevistion(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    int32
		wantErr bool
	}{
		{
			name:    "Empty response",
			output:  "[]",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Valid response",
			output:  `[{"Endpoint":"https://127.0.0.1:2379","Status":{"header":{"cluster_id":14841639068965178418,"member_id":11895648879011905400,"revision":197531225,"raft_term":48069},"version":"3.4.13","dbSize":79740928,"leader":1412153380952468371,"raftIndex":224800270,"raftTerm":48069}}]`,
			want:    197531225,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRevision([]byte(tt.output))
			if (err != nil) != tt.wantErr {
				t.Errorf("getRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRevision() got = %v, want %v", got, tt.want)
			}
		})
	}
}
