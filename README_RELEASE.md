# How to Create the Release

The syncer application has been built for multiple platforms. To create and upload the release:

## Option 1: Using GitHub CLI (Recommended)
If you have GitHub CLI installed:
```bash
bash create_release.sh
```

Or manually:
```bash
gh release create v0.1.0 ./releases/* \
  --title "Syncer v0.1.0" \
  --notes-file RELEASE_NOTES.md
```

## Option 2: Using GitHub Web Interface
1. Create a new tag `v0.1.0`:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0 - Initial syncer release"
   git push origin v0.1.0
   ```

2. Go to: https://github.com/nir414/pc-setup/releases/new?tag=v0.1.0

3. Upload the following files from the `releases/` directory:
   - syncer-windows-amd64.exe
   - syncer-linux-amd64
   - syncer-darwin-amd64
   - syncer-darwin-arm64

4. Copy the content from `RELEASE_NOTES.md` as the release description

5. Mark as "Latest release"

6. Click "Publish release"

## Built Artifacts
All binaries are located in the `releases/` directory and are ready to upload.
