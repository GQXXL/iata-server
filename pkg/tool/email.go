package tool

import (
	"strings"
)

func EmailDomainInWhitelist(email, whitelist string) bool {
	domain := normalizeEmailDomain(emailDomain(email))
	if domain == "" {
		return false
	}

	for _, item := range strings.FieldsFunc(whitelist, isEmailDomainSeparator) {
		allowed := normalizeEmailDomain(item)
		if allowed == "" {
			continue
		}
		if domain == allowed || strings.HasSuffix(domain, "."+allowed) {
			return true
		}
	}

	return false
}

func emailDomain(email string) string {
	atIndex := strings.LastIndex(email, "@")
	if atIndex == -1 || atIndex == len(email)-1 {
		return ""
	}
	return email[atIndex+1:]
}

func normalizeEmailDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "@")
	domain = strings.TrimPrefix(domain, ".")
	return strings.TrimSuffix(domain, ".")
}

func isEmailDomainSeparator(r rune) bool {
	return r == '\n' || r == '\r' || r == ',' || r == ';' || r == ' ' || r == '\t'
}

func MaskEmail(email string) string {
	atIndex := strings.Index(email, "@")
	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return email
	}
	localPart := email[:atIndex]
	domainPart := email[atIndex+1:]

	// 本地部分需要至少保留首字符和末字符
	if len(localPart) < 2 {
		return email
	}
	// 替换本地部分中间字符为星号
	maskedLocal := string(localPart[0]) + strings.Repeat("*", len(localPart)-2) + string(localPart[len(localPart)-1])
	// 返回处理后的邮箱地址
	return maskedLocal + "@" + domainPart
}
