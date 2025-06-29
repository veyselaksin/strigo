# StriGO Documentation

This directory contains the Jekyll-based documentation website for StriGO v2.0.0.

## 🚀 Quick Start

### Local Development

```bash
# Install dependencies
cd docs
bundle install

# Serve locally
bundle exec jekyll serve

# Open http://localhost:4000/strigo/
```

### Docker Development

```bash
# Build and serve with Docker
docker run --rm -v "$PWD/docs:/srv/jekyll" -p 4000:4000 jekyll/jekyll:3.8 jekyll serve
```

## 📁 Structure

```
docs/
├── _config.yaml          # Jekyll configuration
├── index.md              # Homepage (Getting Started)
├── getting-started.md     # Detailed getting started guide
├── api.md                 # Complete API reference
├── advanced.md            # Advanced usage patterns
├── assets/                # Images, CSS, JS
├── _includes/             # Jekyll includes
├── _sass/                 # SCSS files
└── Gemfile               # Ruby dependencies
```

## 🔄 Automated Workflows

### 1. Main Documentation Workflow (`.github/workflows/docs.yml`)

**Triggers:**

- Push to `main` branch (docs changes)
- Pull requests with docs changes
- Manual workflow dispatch

**Features:**

- ✅ Builds Jekyll site
- ✅ Deploys to GitHub Pages
- ✅ Tests links and HTML structure
- ✅ Lighthouse performance audit (PRs)
- ✅ PR comments with preview info

### 2. Manual Documentation Update (`.github/workflows/docs-manual.yml`)

**Usage:**
Go to GitHub Actions → "Manual Documentation Update" → Run workflow

**Options:**

- **Update Type:** patch/minor/major
- **Sync README.md:** Auto-sync from main README
- **Rebuild Benchmarks:** Generate fresh performance charts

**Features:**

- 🔄 Syncs content from README.md
- 📊 Rebuilds benchmark charts
- ⏰ Updates timestamps
- 🧪 Validates documentation
- 📝 Creates update summaries

### 3. README Sync Workflow (`.github/workflows/sync-readme.yml`)

**Triggers:**

- Changes to `README.md` on main branch
- Manual workflow dispatch

**Features:**

- 📥 Auto-extracts sections from README.md
- 🔄 Updates installation, features, performance sections
- ⏰ Adds sync timestamps
- ✅ Validates documentation integrity

## 📝 Content Management

### Updating Documentation

1. **Direct Editing:**

   ```bash
   # Edit files directly
   vim docs/api.md
   git add docs/
   git commit -m "📚 Update API docs"
   git push
   ```

2. **README.md Sync:**

   - Edit main `README.md`
   - Push changes
   - Workflow automatically syncs relevant sections

3. **Manual Workflow:**
   - Use GitHub Actions "Manual Documentation Update"
   - Choose update type and options
   - Workflow handles everything automatically

### Adding New Pages

1. Create new `.md` file in `docs/`
2. Add front matter:
   ```yaml
   ---
   layout: page
   title: Your Page Title
   nav_order: 5
   ---
   ```
3. Update navigation in `_config.yaml` if needed

### Performance Benchmarks

Benchmark charts are generated automatically and stored in:

- `performance_benchmark.png` - Latency comparison
- `throughput_benchmark.png` - Throughput comparison

To regenerate:

1. Use "Manual Documentation Update" workflow
2. Enable "Rebuild performance benchmark charts"
3. Charts will be updated automatically

## 🎨 Theme and Styling

- **Theme:** [Just the Docs](https://just-the-docs.github.io/just-the-docs/)
- **Version:** 0.7.0
- **Customizations:** See `_sass/` directory

### Custom Features

- ✨ Performance benchmark charts integration
- 📊 Code highlighting for Go examples
- 🎯 Custom navigation structure
- 📱 Mobile-responsive design
- 🌙 Dark mode support

## 🧪 Testing

### Local Testing

```bash
cd docs
bundle exec jekyll build
```

### Link Testing

```bash
# Install html-proofer
gem install html-proofer

# Test built site
htmlproofer ./_site --disable-external --check-html
```

### Performance Testing

```bash
# Install Lighthouse CI
npm install -g @lhci/cli

# Build site and run audit
bundle exec jekyll build
lhci autorun
```

## 🚀 Deployment

Documentation is automatically deployed to GitHub Pages:

- **URL:** https://veyselaksin.github.io/strigo/
- **Branch:** `gh-pages` (auto-generated)
- **Trigger:** Push to `main` branch

### Manual Deployment

```bash
# Build and test locally
bundle exec jekyll build
htmlproofer ./_site --disable-external

# Push to trigger deployment
git push origin main
```

## 📊 Analytics and Monitoring

- **GitHub Pages:** Built-in analytics
- **Lighthouse CI:** Performance monitoring
- **HTML Proofer:** Link validation
- **Jekyll Build:** Structure validation

## 🛠️ Troubleshooting

### Common Issues

1. **Bundle install fails:**

   ```bash
   bundle update --bundler
   gem update --system
   ```

2. **Jekyll serve fails:**

   ```bash
   bundle exec jekyll clean
   bundle exec jekyll serve --incremental
   ```

3. **Theme not loading:**

   ```bash
   bundle update just-the-docs
   bundle exec jekyll serve
   ```

4. **GitHub Pages not updating:**
   - Check Actions tab for build errors
   - Verify `_config.yaml` settings
   - Ensure repository has Pages enabled

### Debug Mode

```bash
# Debug Jekyll build
JEKYLL_ENV=development bundle exec jekyll serve --verbose --trace

# Debug with livereload
bundle exec jekyll serve --livereload --incremental
```

## 🔗 Useful Links

- [Jekyll Documentation](https://jekyllrb.com/docs/)
- [Just the Docs Theme](https://just-the-docs.github.io/just-the-docs/)
- [GitHub Pages](https://pages.github.com/)
- [Markdown Guide](https://www.markdownguide.org/)
- [StriGO Repository](https://github.com/veyselaksin/strigo)

## 📄 License

Documentation is licensed under the same terms as the main StriGO project (MIT License).
