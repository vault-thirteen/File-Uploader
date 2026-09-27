package cc

import (
	"path/filepath"

	mime "github.com/vault-thirteen/auxie/MIME"
	cci "github.com/vault-thirteen/auxie/http-helper/CachedContentItem"
)

type CachedContent struct {
	IndexHtml   *cci.CachedContentItem
	ScriptsJs   *cci.CachedContentItem
	Sha256MinJs *cci.CachedContentItem
	StylesCss   *cci.CachedContentItem
	FaviconPng  *cci.CachedContentItem
}

func NewCachedContent(assetsFolderPath string, ttl int) (cc *CachedContent, err error) {
	cc = new(CachedContent)

	cc.IndexHtml, err = cci.NewCachedContentItemFromFile(filepath.Join(assetsFolderPath, Asset_IndexHtml),
		ContentType_Html, ttl)
	if err != nil {
		return nil, err
	}

	cc.ScriptsJs, err = cci.NewCachedContentItemFromFile(filepath.Join(assetsFolderPath, Asset_ScriptsJs), mime.TypeApplicationJavascript, ttl)
	if err != nil {
		return nil, err
	}

	cc.Sha256MinJs, err = cci.NewCachedContentItemFromFile(filepath.Join(assetsFolderPath, Asset_Sha256MinJs), mime.TypeApplicationJavascript, ttl)
	if err != nil {
		return nil, err
	}

	cc.StylesCss, err = cci.NewCachedContentItemFromFile(filepath.Join(assetsFolderPath, Asset_StylesCss), mime.TypeTextCss, ttl)
	if err != nil {
		return nil, err
	}

	cc.FaviconPng, err = cci.NewCachedContentItemFromFile(filepath.Join(assetsFolderPath, Asset_FaviconPng), mime.TypeImagePng, ttl)
	if err != nil {
		return nil, err
	}

	return cc, nil
}
