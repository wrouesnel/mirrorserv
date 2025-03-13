# mirrorserv

Proof of concept web server for on-disk MITM, content-addressed internet emulation.

Basically a server which fakes DNS, TLS, and hostnames to provide an effective all-in-one
hosting solution.

This is a prototype of a larger system.

## DataSource Structure

A bigger system would have a full deduplicating, compressed data source.

This system instead is implemented on top of a barebones filesystem structure that
is easy to read and audit by humans.

### Format

The basic structure is as direct a mapping of web URIs to filesystems as is possible:

The root folder is roughly:
```
<mirror_root>/<backend>/tcp/<port>/<tls|plain>/http/<hostname>/<path>/
```

Underneath the root folder content is stored as:

```
# Marker Files
/current.root : SHA256 reference of the last downloaded content
/current.query.<sha256> : SHA256 reference of the last content downloaded for the hashed URL query path

# Content Files
/content.<sha256> : Content with a matching SHA256 hash
/content.<sha256>.headers : Headers returned by the matching content

# Symlinks
/content : Symlink to the current content file
/content.headers : Symlink to the current content header files
/content.query.<sha256sum>
```