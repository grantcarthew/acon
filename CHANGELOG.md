# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.1] - 2025-12-02

### Added
- Debug command for troubleshooting Markdown conversion
- Goldmark parser support for improved Markdown processing

### Changed
- Improved regex patterns for better document formatting
- Enhanced struct formatting in API and converter packages

### Fixed
- Document table formatting issues in Confluence storage format

## [1.0.0] - 2025-11-27

### Added
- Initial release of acon CLI tool
- Page management commands (create, view, update, delete, list)
- Space management commands (view, list)
- Bidirectional Markdown conversion (Markdown ↔ Confluence storage format)
- Support for Confluence REST API v2
- Environment variable configuration (CONFLUENCE_BASE_URL, CONFLUENCE_EMAIL, CONFLUENCE_API_TOKEN)
- JSON output format support with `-j/--json` flag
- Shell completions (bash, zsh, fish)

[1.0.1]: https://github.com/grantcarthew/acon/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/grantcarthew/acon/releases/tag/v1.0.0
