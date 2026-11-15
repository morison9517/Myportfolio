// =============================================================================
// product_load.go = web/content/products/*.md を読んで、作品の一覧を作る
//
// ▼ 読むタイミング
//
//	本番     … 起動したときに1回だけ。.md を直したら再起動する。
//	開発中   … ページを開くたびに読み直す(保存してブラウザを更新すれば出る)。
//	           ★air は web/ を見張っていないので、こうしないと反映されない。
//
// ▼ ★書き間違いがあったとき
//
//	起動時 … 起動を止めて、どのファイルの何がおかしいかを表示する。
//	         壊れたまま公開されることはない。
//	開発中 … 読み直しに失敗したら、直前に読めた内容のまま表示を続け、
//	         ターミナルに [products] から始まる行を出す。
//
// =============================================================================
package models

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

var (
	productsMu     sync.RWMutex
	loadedProducts []Product

	productDir    string
	productReload bool
)

// SetupProducts = 作品を読み込む。router.go から起動時に1回だけ呼ばれる。
//
//	dir    … .md を置いている場所(web/content/products)
//	reload … true なら、ページを開くたびに読み直す(開発中だけ)
func SetupProducts(dir string, reload bool) error {
	products, err := loadProducts(dir)
	if err != nil {
		return err
	}

	productsMu.Lock()
	loadedProducts = products
	productDir = dir
	productReload = reload
	productsMu.Unlock()

	return nil
}

// AllProducts = 作品の一覧(新しい順)。/products はこれをそのまま出す。
func AllProducts() []Product {
	productsMu.RLock()
	reload := productReload
	productsMu.RUnlock()

	if reload {
		if products, err := loadProducts(productDir); err != nil {
			log.Printf("[products] 読み直せなかったので、前の内容のまま表示します: %v", err)
		} else {
			productsMu.Lock()
			loadedProducts = products
			productsMu.Unlock()
		}
	}

	productsMu.RLock()
	defer productsMu.RUnlock()

	return loadedProducts
}

// loadProducts = フォルダの .md を全部読み、新しい順に並べて返す。
//
// ★README.md は書き方の説明なので読まない。
func loadProducts(dir string) ([]Product, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(files))

	for _, file := range files {
		if strings.EqualFold(filepath.Base(file), "README.md") {
			continue
		}

		product, err := loadProduct(file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(file), err)
		}

		products = append(products, product)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("作品の .md が1つもありません: %s", dir)
	}

	// 新しい順。★同じ年月どうしは作品名の順にして、毎回同じ並びになるようにする。
	sort.SliceStable(products, func(i, j int) bool {
		if products[i].sortKey != products[j].sortKey {
			return products[i].sortKey > products[j].sortKey
		}
		return products[i].Name < products[j].Name
	})

	return products, nil
}

// loadProduct = .md を1つ読んで、作品1件にする。
func loadProduct(file string) (Product, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return Product{}, err
	}

	// ★Windows のメモ帳などで保存すると改行が \r\n になる。
	//   そのままだと下の「--- の行」を見つけられないので、先にそろえる。
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))

	// ★先頭の BOM(UTF-8 の印)も取る。付いていると1行目が --- に見えない。
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))

	front, body, err := splitFrontMatter(raw)
	if err != nil {
		return Product{}, err
	}

	// ★Product に無い項目名を書いたらエラーにする(DisallowUnknownField)。
	//   これが無いと、wide_logo を wideLogo と書き間違えても黙って無視され、
	//   「設定したのに効かない」と悩むことになる。
	var meta struct {
		Product `yaml:",inline"`

		// 年月は "2024-03" で受け取り、下で「2024年3月」に直す。
		Date string `yaml:"date"`
	}

	if err := yaml.UnmarshalWithOptions(front, &meta, yaml.DisallowUnknownField()); err != nil {
		return Product{}, fmt.Errorf("設定欄(--- で囲んだ部分)の書き方が違います:\n%s", yaml.FormatError(err, false, true))
	}

	product := meta.Product
	product.Slug = strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

	if product.Name == "" {
		return Product{}, fmt.Errorf("name がありません")
	}

	if product.Image == "" {
		product.Image = "no-image.svg"
	}

	month, err := time.Parse("2006-01", meta.Date)
	if err != nil {
		return Product{}, fmt.Errorf("date は 2024-03 の形で書いてください(今は %q)", meta.Date)
	}

	product.Date = fmt.Sprintf("%d年%d月", month.Year(), int(month.Month()))
	product.sortKey = month.Format("2006-01")

	html, err := renderMarkdown(body)
	if err != nil {
		return Product{}, fmt.Errorf("本文を変換できませんでした: %w", err)
	}

	product.BodyHTML = html

	return product, nil
}

// splitFrontMatter = 先頭の「--- から --- まで」と、それより下の本文に分ける。
func splitFrontMatter(raw []byte) (front, body []byte, err error) {
	const fence = "---\n"

	if !bytes.HasPrefix(raw, []byte(fence)) {
		return nil, nil, fmt.Errorf("1行目は --- にしてください(設定欄の始まり)")
	}

	rest := raw[len(fence):]

	// 閉じの --- は行頭にあるものだけを探す。本文中の --- (区切り線)と取り違えないため。
	// ★最初に見つかったものが閉じ。本文の区切り線はそれより下にある。
	if front, body, found := bytes.Cut(rest, []byte("\n"+fence)); found {
		return front, body, nil
	}

	// ファイルの最後が改行なしの --- で終わっている場合(本文なし)。
	if front, found := bytes.CutSuffix(rest, []byte("\n---")); found {
		return front, nil, nil
	}

	return nil, nil, fmt.Errorf("設定欄の終わりの --- が見つかりません")
}
