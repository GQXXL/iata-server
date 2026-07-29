package tool

import "testing"

func TestEmailDomainInWhitelist(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		whitelist string
		want      bool
	}{
		{
			name:      "exact domain",
			email:     "user@qq.com",
			whitelist: "gmail.com\nqq.com",
			want:      true,
		},
		{
			name:      "case insensitive",
			email:     "USER@GMAIL.COM",
			whitelist: "gmail.com",
			want:      true,
		},
		{
			name:      "comma separated",
			email:     "user@outlook.com",
			whitelist: "gmail.com, outlook.com",
			want:      true,
		},
		{
			name:      "subdomain",
			email:     "user@mail.example.com",
			whitelist: "example.com",
			want:      true,
		},
		{
			name:      "suffix without dot boundary is rejected",
			email:     "user@badexample.com",
			whitelist: "example.com",
			want:      false,
		},
		{
			name:      "empty whitelist",
			email:     "user@qq.com",
			whitelist: "",
			want:      false,
		},
		{
			name:      "invalid email",
			email:     "not-an-email",
			whitelist: "example.com",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmailDomainInWhitelist(tt.email, tt.whitelist); got != tt.want {
				t.Fatalf("EmailDomainInWhitelist(%q, %q) = %v, want %v", tt.email, tt.whitelist, got, tt.want)
			}
		})
	}
}
