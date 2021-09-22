# Changelog

## Next 
### New and Improved
- Filter map fields which are not tagged. ([PR](https://github.com/hashicorp/go-eventlogger/pull/63))
- Setting filters to no operation. ([PR](https://github.com/hashicorp/go-eventlogger/pull/61))

##  filters/encrypt/v0.1.3 (2021/08/23)
### New and Improved
- Support tagged fields and taggable ([PR](https://github.com/hashicorp/go-eventlogger/pull/60))

## filters/encrypt/v0.1.2 (2021/08/16)
### New and Improved
- feature (encrypt): make the wrapper optional, unless a configured filter
  operation requires it ([PR](https://github.com/hashicorp/go-eventlogger/pull/59))

## filters/encrypt/v0.1.1 (2021/08/10)
### New and Improved
- Add encrypt.DefaultFilterOperations(...) so clients can easily get the
  defaults without splunking the code base and/or creating some sort of shadow
  enums in their code bases
  ([PR](https://github.com/hashicorp/go-eventlogger/pull/58))

## filters/encrypt/v0.1.0 (2021/07/27)
### New and Improved
- Update filters/encrypt deps to latest tagged eventlogger ([PR](https://github.com/hashicorp/go-eventlogger/pull/56))
- Reduce mod dependencies ([PR](https://github.com/hashicorp/go-eventlogger/pull/55))
- Shorten classification tag to just class ([PR](https://github.com/hashicorp/go-eventlogger/pull/48))
- Clean up some of the godocs for packages and add more examples ([PR](https://github.com/hashicorp/go-eventlogger/pull/47))
- Add a new encrypt package which implements a new Filter node. ([PR](https://github.com/hashicorp/go-eventlogger/pull/46))