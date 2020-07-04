package item

import (
	"crypto/sha256"
	"hash"
	"hash/crc32"
	"strings"

	"github.com/progrhyme/binq/internal/logs"
)

type ItemChecksum struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256,omitempty"`
	// CRC-32 IEEE Std.
	CRC string `json:"crc,omitempty"`
}

func NewItemChecksums(arg string) (sums []ItemChecksum) {
	if arg == "" {
		return nil
	}

	for _, entry := range strings.Split(arg, ",") {
		params := strings.Split(entry, ":")
		switch len(params) {
		case 2:
			sums = append(sums, ItemChecksum{File: params[0], SHA256: params[1]})
		case 3:
			switch params[2] {
			case "sha256", "SHA256", "SHA-256":
				sums = append(sums, ItemChecksum{File: params[0], SHA256: params[1]})
			case "crc", "CRC":
				sums = append(sums, ItemChecksum{File: params[0], CRC: params[1]})
			default:
				logs.Warnf("Unsupported algorithm: %s. Param: %s", params[2], entry)
			}

		default:
			logs.Warnf("Wrong argement for replacement: %s", entry)
		}
	}
	return sums
}

func (rev *ItemRevision) GetChecksum(file string) (sum *ItemChecksum) {
	for _, cs := range rev.Checksums {
		if cs.File == file {
			return &cs
		}
	}
	return nil
}

func (sum *ItemChecksum) GetSumAndHasher() (s string, h hash.Hash) {
	if sum.SHA256 != "" {
		return sum.SHA256, sha256.New()
	} else if sum.CRC != "" {
		return sum.CRC, crc32.NewIEEE()
	}
	return "", nil
}
