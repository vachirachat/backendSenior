package service

import (
	"common/utils/ginutils"
	"fmt"
	"github.com/go-resty/resty/v2"
)

type StickerService struct {
	c *resty.Client
}

func NewStickerService(basePath string) *StickerService {
	clnt := resty.New()
	clnt.SetHostURL(basePath)
	return &StickerService{
		c: clnt,
	}
}

func (s *StickerService) CheckSticker(ID string) (bool, error) {
	var result ginutils.Response
	if _, err := s.c.R().SetResult(&result).Get("/api/v1/sticker/check/" + ID); err != nil {
		return false, fmt.Errorf("request error : %w", err)
	}
	return result.Data.(bool), nil
}
