// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/k14s/imgpkg/pkg/imgpkg/imagedesc"
	"github.com/k14s/imgpkg/pkg/imgpkg/imagetar"

	regname "github.com/google/go-containerregistry/pkg/name"
	ctlbundle "github.com/k14s/imgpkg/pkg/imgpkg/bundle"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"
	ctlimgset "github.com/k14s/imgpkg/pkg/imgpkg/imageset"
	"github.com/k14s/imgpkg/pkg/imgpkg/lockconfig"
	"github.com/k14s/imgpkg/pkg/imgpkg/plainimage"
)

type CopyRepoSrc struct {
	ImageFlags              ImageFlags
	BundleFlags             BundleFlags
	LockInputFlags          LockInputFlags
	IncludeNonDistributable bool
	Concurrency             int
	logger                  *ctlimg.LoggerPrefixWriter
	imageSet                ctlimgset.ImageSet
	tarImageSet             ctlimgset.TarImageSet
	registry                ctlimgset.ImagesReaderWriter
}

func (o CopyRepoSrc) CopyToTar(dstPath string) error {
	unprocessedImageRefs, err := o.getSourceImages()
	if err != nil {
		return err
	}

	ids, err := o.tarImageSet.Export(unprocessedImageRefs, dstPath, o.registry, imagetar.NewImageLayerWriterCheck(o.IncludeNonDistributable))
	if err != nil {
		return err
	}

	informUserToUseTheNonDistributableFlagWithDescriptors(o.logger, o.IncludeNonDistributable, imageRefDescriptorsMediaTypes(ids))

	return nil
}

func (o CopyRepoSrc) CopyToRepo(repo string) (*ctlimgset.ProcessedImages, error) {
	unprocessedImageRefs, err := o.getSourceImages()
	if err != nil {
		return nil, err
	}

	importRepo, err := regname.NewRepository(repo)
	if err != nil {
		return nil, fmt.Errorf("Building import repository ref: %s", err)
	}

	processedImages, ids, err := o.imageSet.Relocate(unprocessedImageRefs, importRepo, o.registry)
	if err != nil {
		return nil, err
	}

	informUserToUseTheNonDistributableFlagWithDescriptors(o.logger, o.IncludeNonDistributable, imageRefDescriptorsMediaTypes(ids))

	return processedImages, nil
}

func (o CopyRepoSrc) getSourceImages() (*ctlimgset.UnprocessedImageRefs, error) {
	unprocessedImageRefs := ctlimgset.NewUnprocessedImageRefs()

	switch {
	case o.LockInputFlags.LockFilePath != "":
		bundleLock, imagesLock, err := lockconfig.NewLockFromPath(o.LockInputFlags.LockFilePath)
		if err != nil {
			return nil, err
		}

		switch {
		case bundleLock != nil:
			_, imageRefs, err := o.getBundleImageRefs(bundleLock.Bundle.Image)
			if err != nil {
				return nil, err
			}

			for _, img := range imageRefs {
				unprocessedImageRefs.Add(ctlimgset.UnprocessedImageRef{DigestRef: img.PrimaryLocation()})
			}

			unprocessedImageRefs.Add(ctlimgset.UnprocessedImageRef{
				DigestRef: bundleLock.Bundle.Image,
				Tag:       bundleLock.Bundle.Tag,
			})

			return unprocessedImageRefs, nil

		case imagesLock != nil:
			for _, img := range imagesLock.Images {
				plainImg := plainimage.NewPlainImage(img.Image, o.registry)

				ok, err := ctlbundle.NewBundleFromPlainImage(plainImg, o.registry).IsBundle()
				if err != nil {
					return nil, err
				}
				if ok {
					return nil, fmt.Errorf("Unable to copy bundles using an Images Lock file (hint: Create a bundle with these images)")
				}

				unprocessedImageRefs.Add(ctlimgset.UnprocessedImageRef{DigestRef: plainImg.DigestRef()})
			}
			return unprocessedImageRefs, nil

		default:
			panic("Unreachable")
		}

	case o.ImageFlags.Image != "":
		plainImg := plainimage.NewPlainImage(o.ImageFlags.Image, o.registry)

		ok, err := ctlbundle.NewBundleFromPlainImage(plainImg, o.registry).IsBundle()
		if err != nil {
			return nil, err
		}
		if ok {
			return nil, fmt.Errorf("Expected bundle flag when copying a bundle (hint: Use -b instead of -i for bundles)")
		}

		unprocessedImageRefs.Add(ctlimgset.UnprocessedImageRef{DigestRef: plainImg.DigestRef()})
		return unprocessedImageRefs, nil

	default:
		bundle, imageRefs, err := o.getBundleImageRefs(o.BundleFlags.Bundle)
		if err != nil {
			return nil, err
		}

		for _, img := range imageRefs {
			unprocessedImageRefs.Add(ctlimgset.UnprocessedImageRef{DigestRef: img.PrimaryLocation()})
		}

		unprocessedImageRefs.Add(ctlimgset.UnprocessedImageRef{DigestRef: bundle.DigestRef(), Tag: bundle.Tag()})

		return unprocessedImageRefs, nil
	}

	panic("Unreachable")
}

func (o CopyRepoSrc) getBundleImageRefs(bundleRef string) (*ctlbundle.Bundle, []lockconfig.ImageRef, error) {
	bundle := ctlbundle.NewBundle(bundleRef, o.registry)

	imgLock, err := bundle.AllImagesLock(o.Concurrency)
	if err != nil {
		if ctlbundle.IsNotBundleError(err) {
			return nil, nil, fmt.Errorf("Expected bundle image but found plain image (hint: Did you use -i instead of -b?)")
		}
		return nil, nil, err
	}

	imageRefs, err := imgLock.LocationPrunedImageRefs()
	if err != nil {
		return nil, nil, fmt.Errorf("Pruning image ref locations: %s", err)
	}
	return bundle, imageRefs, nil
}

func imageRefDescriptorsMediaTypes(ids *imagedesc.ImageRefDescriptors) []string {
	mediaTypes := []string{}
	for _, descriptor := range ids.Descriptors() {
		if descriptor.Image != nil {
			for _, layerDescriptor := range (*descriptor.Image).Layers {
				mediaTypes = append(mediaTypes, layerDescriptor.MediaType)
			}
		}

	}
	return mediaTypes
}
