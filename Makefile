test:
	@CGO_CXXFLAGS="$$(go env CGO_CXXFLAGS) -Wno-deprecated" go test -v

install:
	@CGO_CXXFLAGS="$$(go env CGO_CXXFLAGS) -Wno-deprecated" go install ./cmd/zetasqlfmt
