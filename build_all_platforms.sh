#!/usr/bin/env bash

platforms=("linux/amd64" "linux/386" "linux/arm" "linux/arm64" "linux/mips" "linux/mips64" "darwin/arm64" "darwin/amd64" "freebsd/amd64" "freebsd/386" "windows/386" "windows/amd64")
thedir=builds
thedir_tarballs=$thedir/tarballs
rm -rf $thedir
mkdir -p $thedir_tarballs

package=preconnect_balproxy.go
package_name=preconnect_balproxy

go build -o $thedir/preconnect_balproxy  preconnect_balproxy.go

for platform in "${platforms[@]}"
do	
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$package_name'-'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi	

	env GOOS=$GOOS GOARCH=$GOARCH go build -o $thedir/$output_name $package
	tar cvfz $thedir_tarballs/$output_name.tar.gz $thedir/$output_name
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done

package=after_connect_version/after_connect_balproxy.go
package_name=after_connect_balproxy

go build -o $thedir/after_connect_balproxy after_connect_version/after_connect_balproxy.go

for platform in "${platforms[@]}"
do	
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=$package_name'-'$GOOS'-'$GOARCH
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi	

	env GOOS=$GOOS GOARCH=$GOARCH go build -o $thedir/$output_name $package
	tar cvfz $thedir_tarballs/$output_name.tar.gz $thedir/$output_name
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done


