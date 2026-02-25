# CLAUDE.md

このファイルは、リポジトリ内で作業する Claude Code (claude.ai/code) へのガイダンスを提供します。

## よく使うコマンド

```bash
# ビルド
go build ./...

# テスト（race 検出器なし — Windows で CGO なしでも安全）
go test ./...

# race 検出器付きテスト（Windows では CGO_ENABLED=1 が必要。CI の Linux では不要）
go test -v -race ./...

# 特定のパッケージ・テストだけ実行
go test ./internal/crawler/...
go test -run TestInScope ./internal/crawler/

# Lint（golangci-lint のインストールが必要）
golangci-lint run
```

## アーキテクチャ

Cobra を通じて CLI サブコマンドとして公開された 3 段階パイプライン:

1. **`crawler`** (`internal/crawler/`) — BFS ウェブクローラー。メインゴルーチンのディスパッチャーがバッファ付き `jobs` チャネルを通じて URL をワーカープールに流す。ワーカーはページを取得し `results` チャネルに結果を送る。`urlutil.go` が URL の正規化・スコープ判定（`InScope`）・リンク抽出・URL→ファイル名変換を担当。

2. **`converter`** (`internal/converter/`) — HTML→Markdown 変換器。入力ディレクトリを走査（または ZIP を展開）し、セマフォ型ワーカープールで並列処理する。各ファイルは `ExtractMainNode`（goquery セレクターで最適なコンテンツノードを取得）→ `Transform`（不要タグ/クラス/画像を除去後、`html-to-markdown` の `ConvertString` を呼び出す）の順に処理される。

3. **`combiner`** (`internal/combiner/`) — Markdown 結合器。`WalkDir` で辞書順に `.md` ファイルを収集し、単一の出力ファイルに結合する。総語数が `MaxWords`（デフォルト 50 万語）を超える場合、番号付きファイル（`combined-001.md`, `combined-002.md`, …）に自動分割する。

`pipeline` サブコマンドは 3 つを順番に実行し、各ステージの出力ディレクトリを次のステージに渡す。

## コミット規約

[Conventional Commits](https://www.conventionalcommits.org/) に従う。

```
<type>(<scope>): <summary>
```

**type（必須）**

| type | 用途 |
|---|---|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメントのみの変更 |
| `refactor` | 機能追加・バグ修正を伴わないコード変更 |
| `test` | テストの追加・修正 |
| `chore` | ビルド・CI・依存関係などの雑務 |
| `perf` | パフォーマンス改善 |

**scope（任意）** — 変更対象のパッケージ名など（例: `crawler`, `converter`, `combiner`, `cmd`）

**破壊的変更** — フッターに `BREAKING CHANGE: <説明>` を追加するか、type/scope の直後に `!` を付ける（例: `feat!: ...`）

**GitHub Issue 関連コミット** — Issue に紐づく場合はトレーラーを追加:

```
git commit --trailer "Github-Issue:#<number>"
```

これにより、コミットメッセージ末尾に以下のトレーラーが付与される:

```
Github-Issue: #<number>
```

**ツール帰属の禁止** — `Co-Authored-By:` や `Co-authored-by:`、その他使用ツールへの
言及をコミットメッセージ・PR 本文に含めないこと。

### 例

```
feat(crawler): add --max-depth flag to limit BFS depth
fix(converter): handle empty HTML body without panic
docs: translate CLAUDE.md to Japanese
chore(ci): upgrade Go version to 1.24
feat!: change combine output format

BREAKING CHANGE: combined output files now use 0-indexed numbering
```

## 注意事項

- **`goquery.NewDocumentFromNode`** は戻り値が 1 つ（エラーなし）。`NewDocumentFromReader` とは異なるため、第 2 戻り値を追加しないこと。
- **`InScope`** のスコープ定義: 同一ホスト かつ ターゲットパスが `base.Path + "/"` をプレフィックスに持つ（またはベースパスと完全一致）。末尾スラッシュの処理は `urlutil.go:InScope` にある。
- ロギングには `log/slog` を全体で使用。デフォルトレベルは `WARN`; `--verbose` で `DEBUG` に切り替わる。
- オプションの `--report` フラグは JSON レポート（`internal/report/`）を書き出す。レポートに記録されるのは `crawl` と `convert` ステップのみ（`combine` は対象外）。
- テストは全ファイル I/O に `t.TempDir()` を使用。永続的なフィクスチャはない。
