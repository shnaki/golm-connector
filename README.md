# golm-connector

> ウェブサイトを BFS クロールし、HTML を Markdown に変換して単一ドキュメントに結合するツール

![CI](https://github.com/shnaki/golm-connector/actions/workflows/ci.yml/badge.svg)

## 概要

`golm-connector` は 3 段階のパイプラインで動作する CLI ツールです。

1. **crawl** — BFS（幅優先探索）でウェブサイトを巡回し、HTML ファイルとして保存する
2. **convert** — 保存された HTML ファイルをコンテンツ抽出しながら Markdown に変換する
3. **combine** — 変換された Markdown ファイルを 1 つのドキュメントに結合する（語数上限で自動分割）

[NotebookLM](https://notebooklm.google/) などへのソース投入用ドキュメント作成を想定しています。

## インストール

Go 1.23 以上が必要です。

```bash
git clone https://github.com/shnaki/golm-connector.git
cd golm-connector
go build -o golm-connector .
```

ビルドした `golm-connector` バイナリを `PATH` の通ったディレクトリに配置してください。

## 使い方

### クイックスタート（pipeline コマンド）

クロール・変換・結合を一括実行するには `pipeline` サブコマンドを使います。

```bash
golm-connector pipeline https://example.com/docs/ -o out/
```

実行後、以下のディレクトリ構成で出力されます。

```
out/
├── html/          # クロールで保存された HTML ファイル
├── md/            # 変換された Markdown ファイル
└── combined.md    # 結合済みドキュメント（語数超過時は combined-001.md, combined-002.md, ...）
```

### 個別サブコマンド

#### crawl

指定 URL を起点に BFS でクロールし、同一ホスト・同一パス配下のページを HTML として保存します。

```bash
golm-connector crawl https://example.com/docs/ -o html_output/
```

| フラグ | デフォルト | 説明 |
|---|---|---|
| `-o / --output` | `html_output` | HTML 保存先ディレクトリ |
| `--max-pages` | `0`（無制限） | クロールする最大ページ数 |
| `--delay` | `1s` | リクエスト間の待機時間（例: `500ms`, `2s`） |
| `--max-concurrency` | `5` | 並列 HTTP ワーカー数 |
| `--cache-dir` | `""` | HTTP レスポンスのディスクキャッシュ先 |
| `--retry-from-report` | `""` | 前回 `--report` で出力した JSON の失敗 URL を再試行 |

#### convert

HTML ファイルのディレクトリ（または ZIP）を受け取り、Markdown に変換します。

```bash
golm-connector convert html_output/ -o md_output/

# ZIP ファイルを直接変換する場合
golm-connector convert . --zip archive.zip -o md_output/
```

| フラグ | デフォルト | 説明 |
|---|---|---|
| `-o / --output` | `md_output` | Markdown 出力先ディレクトリ |
| `--zip` | `""` | HTML を含む ZIP ファイルのパス |
| `--max-workers` | `4` | 並列変換ワーカー数 |
| `--strip-tags` | `""` | 削除する HTML タグ（カンマ区切り、例: `nav,footer`） |
| `--strip-classes` | `""` | 削除する CSS クラス（カンマ区切り） |
| `--retry-from-report` | `""` | 前回 `--report` で出力した JSON の失敗ファイルを再試行 |

#### combine

Markdown ファイルのディレクトリを受け取り、辞書順に 1 ファイルへ結合します。
語数が `--max-words` を超える場合は `combined-001.md`, `combined-002.md`, ... と自動分割します。

```bash
golm-connector combine md_output/ -o combined.md
```

| フラグ | デフォルト | 説明 |
|---|---|---|
| `-o / --output` | `combined.md` | 出力ファイルパス |
| `--max-words` | `500000` | 出力ファイルあたりの最大語数（`0` で無制限） |

#### pipeline

`crawl → convert → combine` を順番に実行します。

```bash
golm-connector pipeline https://example.com/docs/ -o out/ \
  --max-pages 200 \
  --delay 500ms \
  --strip-tags nav,footer,aside
```

| フラグ | デフォルト | 説明 |
|---|---|---|
| `-o / --output` | `pipeline_output` | ベース出力ディレクトリ |
| `--max-pages` | `0` | 最大クロールページ数 |
| `--delay` | `1s` | クロールリクエスト間の待機時間 |
| `--max-concurrency` | `5` | 並列クロールワーカー数 |
| `--max-workers` | `4` | 並列変換ワーカー数 |
| `--strip-tags` | `""` | 削除する HTML タグ |
| `--strip-classes` | `""` | 削除する CSS クラス |
| `--max-words` | `500000` | 出力ファイルあたりの最大語数 |

### グローバルフラグ

すべてのサブコマンドで使用できます。

| フラグ | 説明 |
|---|---|
| `-v / --verbose` | デバッグログを stderr に出力する |
| `--report <path>` | crawl・convert の結果を JSON ファイルに書き出す |

### JSON レポート

`--report` フラグを指定すると、crawl と convert の処理結果（成功 URL・失敗 URL とエラー内容）を JSON ファイルに記録します。
失敗した URL やファイルを `--retry-from-report` で再試行する際に使用できます。

```bash
# 最初の実行（レポートを保存）
golm-connector crawl https://example.com/docs/ -o html/ --report report.json

# 失敗した URL のみ再試行
golm-connector crawl https://example.com/docs/ -o html/ --retry-from-report report.json
```

## 開発

```bash
# ビルド
go build .

# テスト
go test ./...

# Lint（golangci-lint が必要）
golangci-lint run
```
