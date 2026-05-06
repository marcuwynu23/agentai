# AgentAI v1.1.0 Release Notes

**Release Date:** May 6, 2026

## Overview

This release focuses on documentation improvements, workflow optimizations, and enhanced user experience with better release management and CI/CD improvements.

## What's New

### Documentation Enhancements
- **Professional USAGE.md**: Complete rewrite with TUI-focused documentation
- **Comprehensive Architecture Docs**: Updated technical documentation reflecting current codebase
- **Enhanced Contributing Guide**: Project-specific development guidelines
- **Release Documentation**: Structured release notes for each version

### Workflow Improvements
- **Direct GitHub Releases**: Eliminated artifact dependencies for faster releases
- **Node.js 24 Compatibility**: Updated all GitHub Actions to use compatible versions
- **Optimized Release Assets**: Consistent naming pattern with version prefixes
- **Cleaner CI/CD**: Streamlined build and deployment process

### User Experience
- **Better Asset Naming**: Release assets now follow `agentai-{VERSION}-{OS}-{ARCH}` pattern
- **Improved Documentation Structure**: Clear navigation and comprehensive guides
- **Enhanced Troubleshooting**: Better error handling and support documentation

## Changes

### Documentation Structure
- **Reorganized docs/**: Added version-specific release notes in `docs/releases/`
- **Updated README.md**: Added comprehensive documentation references
- **Improved CHANGELOG.md**: More user-friendly format with better context
- **Enhanced Templates**: Better release note templates for future versions

### Workflow Optimizations
- **Removed Deprecated Actions**: Eliminated Node.js 20 deprecation warnings
- **Direct Publishing**: Assets now publish directly to GitHub releases
- **Better Graph Display**: Fixed GitHub Actions workflow visualization
- **Consistent Naming**: Standardized release asset names across platforms

## Fixes

### GitHub Actions
- **Fixed Workflow Graph**: Resolved "graph cannot be shown" error
- **Eliminated Deprecation Warnings**: Updated to Node.js 24 compatible actions
- **Improved Release Process**: Fixed asset naming inconsistencies

### Documentation
- **Fixed TUI References**: Updated all documentation to reflect TUI nature
- **Corrected Provider Information**: Added Cloudflare AI Gateway support
- **Enhanced Examples**: Better usage examples and troubleshooting guides

## Performance Improvements

### Build Process
- **Faster Releases**: Direct publishing eliminates artifact download/upload delays
- **Cleaner Workflows**: Fewer steps and reduced complexity
- **Better Error Handling**: Improved GitHub Actions reliability

### Documentation
- **Faster Navigation**: Better organized documentation structure
- **Clearer Instructions**: User-focused language and examples
- **Comprehensive Coverage**: Complete documentation for all features

## Platform Support

### Supported Platforms (Unchanged)
- **Linux**: amd64, arm64
- **Windows**: amd64, arm64  
- **macOS**: amd64, arm64

### Asset Naming
- **Linux**: `agentai-v1.1.0-linux-amd64.tar.gz`
- **Windows**: `agentai-v1.1.0-windows-amd64.zip`
- **macOS**: `agentai-v1.1.0-darwin-amd64.tar.gz`

## Technical Details

### GitHub Actions Updates
- **actions/checkout@v4**: Maintained (Node.js 24 compatible)
- **actions/setup-go@v5**: Maintained (Node.js 24 compatible)
- **softprops/action-gh-release@v2**: Maintained (Node.js 24 compatible)
- **Removed**: actions/upload-artifact@v4, actions/download-artifact@v4

### Documentation Structure
```
docs/
├── releases/
│   ├── v1.0.0/
│   │   └── RELEASE-NOTES.md
│   └── v.1.1.0/
│       └── RELEASE-NOTES.md
├── architecture.md
└── GUIDE.md
```

## Documentation Links

### Release Notes
- **[v1.0.0 Release Notes](v1.0.0/RELEASE-NOTES.md)**: Initial release information
- **[Main Documentation](../../README.md)**: Project overview and getting started
- **[Usage Guide](../../USAGE.md)**: Complete usage instructions
- **[Contributing Guide](../../CONTRIBUTING.md)**: Development guidelines

### Technical Documentation
- **[Architecture](../architecture.md)**: System design and technical details
- **[CHANGELOG.md](../../CHANGELOG.md)**: Version history and changes

## Migration Guide from v1.0.0

### For Users
No breaking changes. This is a documentation and workflow improvement release.

### For Developers
- **Documentation Structure**: New version-specific release notes in `docs/releases/`
- **GitHub Actions**: Updated workflow may require review if you have custom workflows
- **Asset Naming**: Downloaded assets now include version in filename

## Known Issues

### Resolved in This Release
- GitHub Actions deprecation warnings
- Workflow graph display issues
- Inconsistent release asset naming

### Current Limitations (Unchanged)
- Requires internet connection for AI providers
- Local Ollama setup required for offline usage

## What's Next

### Planned for v1.2.0
- Enhanced error handling and recovery
- Additional AI provider support
- Docker containerization
- Plugin system for custom providers
- Improved project templates

### Long-term Roadmap
- Web interface
- Team collaboration features
- Advanced project analysis
- Integration with popular IDEs

## Contributing

Contributions welcome! See our [Contributing Guidelines](../../CONTRIBUTING.md) for:
- Code style and conventions
- Pull request process
- Issue reporting
- Feature requests

## License

This release is licensed under the MIT License. See [LICENSE](../../LICENSE) for full details.

---

**Thank you for using AgentAI v1.1.0!**

For questions, issues, or suggestions, please visit our [GitHub repository](https://github.com/marcuwynu23/agentai).
