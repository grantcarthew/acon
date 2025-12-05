# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.0](https://github.com/grantcarthew/acon/compare/v1.1.0...v1.2.0) - 2025-12-05

### Added

- Task list support in Markdown conversion (`- [ ]` and `- [x]`)
- Strikethrough support (`~~text~~`)
- External image support in Markdown to Confluence conversion
- Comprehensive test data and round-trip testing script (`testdata/`)
- Debug command documentation in README

### Changed

- Improved code block handling, including empty blocks and special characters
- Tighter nested list formatting

### Fixed

- Over-escaping issues in generated Markdown output

## [1.1.0](https://github.com/grantcarthew/acon/compare/v1.0.1...v1.1.0) - 2025-12-04

### Added

- `page move` command to move pages to a new parent
- `--parent` flag for `page list` to list child pages
- `--sort` and `--desc` flags for `page list` sorting options
- Confluence code macro preprocessing for StorageToMarkdown conversion

### Changed

- Updated Go module dependencies

### Fixed

- HTML entity decoding in Markdown output
- Error handling improvements in page and space commands

## [1.0.1](https://github.com/grantcarthew/acon/compare/v1.0.0...v1.0.1) - 2025-12-02

### Added

- Debug command for troubleshooting Markdown conversion
- Goldmark parser support for improved Markdown processing

### Changed

- Improved regex patterns for better document formatting
- Enhanced struct formatting in API and converter packages

### Fixed

- Document table formatting issues in Confluence storage format

## <https://github.com/grantcarthew/acon/releases/tag/v1.0.0> - 2025-11-27

### Added

- Initial release of acon CLI tool
- Page management commands (create, view, update, delete, list)
- Space management commands (view, list)
- Bidirectional Markdown conversion (Markdown ↔ Confluence storage format)
- Support for Confluence REST API v2
- Environment variable configuration (CONFLUENCE_BASE_URL, CONFLUENCE_EMAIL, CONFLUENCE_API_TOKEN)
- JSON output format support with `-j/--json` flag
- Shell completions (bash, zsh, fish)

