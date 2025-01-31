blog
====

<p align="center">
  <img src="screenshot.png" width="600">
</p>

This tool is mainly for writing [tellme.tokyo](https://github.com/babarot/tellme.tokyo) blog. If you run this app with `edit` command, you can choise the exisiting articles with cool [prompt](https://github.com/manifoldco/promptui). Internally, this app makes hugo server run as background process. So it makes you easy to access localhost:1313 while writing your post without calling hugo server separately!

## Installation

Download the binary from [GitHub Releases][release] and drop it in your `$PATH`.

- [Darwin / Mac][release]
- [Linux][release]

## Todos

- [ ] Update readme, add demo
- [ ] Add `new` command
  - [ ] asking title (input one line)
  - [ ] asking date (date picker)
  - [ ] asking other metadata
- [ ] add help model, key binds

## License

[MIT][license]


[release]: https://github.com/babarot/blog/releases/latest
[license]: https://babarot.mit-license.org
