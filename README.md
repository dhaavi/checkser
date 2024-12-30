# Checkser

Have you ever wondered if your file storage solution really checks if your files have been corrupted? Or some of them went missing? Or a big data transfer went weird, you had to retry a couple times and are not sure anymore if files are _really_ correct and complete?

With `checkser`, you can easily verify if everything is as it should be.

It adds a simple `.checkser.yaml` file to each dir, that looks something like this:

```
checkser: 1
updated_at: 2024-12-30T14:59:42Z
updated_by: vps-349785
files:
    - name: Dockerfile
      size: 357
      mod: 2024-12-30T13:40:45.781172594Z
      alg: BLAKE2b_256
      sum: 2a8f864abd0baf8933e99f072db393cd81048db5a6d335700560696b8694eabc
    - name: docker-compose.yml
      size: 382
      mod: 2024-12-30T13:40:45.785172594Z
      alg: BLAKE2b_256
      sum: 09daccc2dacdae05eba12a9e1ab1edbda75279f2e98e009d5d9336c3c9520371
    - name: style.css
      size: 407
      mod: 2024-12-30T13:40:45.788172594Z
      alg: BLAKE2b_256
      sum: 45ab72530941d764659830aa4f3f304b8274f3a98f1f1fffada776ea38a253f4
dirs:
    - name: _site
      alg: BLAKE2b_256
      sum: 3e6d71b0faef106908920cd77bf369d874314e804dd559af123e0d9c9e4ecb51
    - name: jekyll
      alg: BLAKE2b_256
      sum: 42881642126438f0290925a9829ef5ecfe68f8a1b21fa4ae82b474dc3655ecff
```

`checkser` is only concerned with integrity, but as all checksum files have checksums in the parent checksum, you can simply sign the root checksum file with `age` or whatever you fancy.

## Usage

- `checkser /tmp/test` Interactive mode.
- `checkser update /tmp/test` Update checksum files to reflect file system changes. (non-interactive)
- `checkser verify /tmp/test` Verify all checksums. (non-interactive)
