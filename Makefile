test:
	@CGO_CXXFLAGS="$$(go env CGO_CXXFLAGS) -Wno-deprecated" go test -v
