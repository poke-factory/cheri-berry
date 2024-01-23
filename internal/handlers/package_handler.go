package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/poke-factory/cheri-berry/internal/requests"
	"github.com/poke-factory/cheri-berry/internal/storage"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type PackageJson struct {
	Name     string                    `json:"name"`
	Versions map[string]PackageVersion `json:"versions"`
	Time     map[string]time.Time      `json:"time"`
	Users    struct {
	} `json:"users"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Distfiles struct {
		Demo001Tgz struct {
			Url string `json:"url"`
			Sha string `json:"sha"`
		} `json:"demo-0.0.1.tgz"`
		Demo0110Tgz struct {
			Url string `json:"url"`
			Sha string `json:"sha"`
		} `json:"demo-0.1.10.tgz"`
	} `json:"_distfiles"`
	Attachments map[string]struct {
		Shasum string `json:"shasum"`
	} `json:"_attachments"`
	Rev    string `json:"_rev"`
	Readme string `json:"readme"`
	Id     string `json:"_id"`
}

type PackageVersion struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Main        string `json:"main"`
	Scripts     struct {
		Test string `json:"test"`
	} `json:"scripts"`
	Author struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	License     string `json:"license"`
	Id          string `json:"_id"`
	NodeVersion string `json:"_nodeVersion"`
	NpmVersion  string `json:"_npmVersion"`
	Dist        struct {
		Integrity string `json:"integrity"`
		Shasum    string `json:"shasum"`
		Tarball   string `json:"tarball"`
	} `json:"dist"`
	Contributors []interface{} `json:"contributors"`
}

func UploadPackage(c *gin.Context) {
	var request requests.UploadPackageRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for filename := range request.Attachments {
		packageFile := fmt.Sprintf("cheri-berry/%s/%s", request.Name, filename)

		fileExist, err := storage.Storage.Exists(context.TODO(), packageFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if fileExist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "version already exists"})
			return
		}
	}

	packageJsonFile := fmt.Sprintf("cheri-berry/%s/package.json", request.Name)

	exists, err := storage.Storage.Exists(context.TODO(), packageJsonFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pkgJson PackageJson

	if exists {
		content, err := storage.Storage.GetBytes(context.TODO(), packageJsonFile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = json.Unmarshal(content, &pkgJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		pkgJson = convertPkgJson(request, &pkgJson)
	} else {
		pkgJson = convertPkgJson(request, nil)
	}
	pkgJsonStr, err := json.Marshal(pkgJson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r := bytes.NewReader(pkgJsonStr)
	err = storage.Storage.Put(context.TODO(), packageJsonFile, r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for filename, attachment := range request.Attachments {
		data, err := base64.StdEncoding.DecodeString(attachment.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		reader := bytes.NewReader(data)

		err = storage.Storage.Put(context.TODO(), fmt.Sprintf("cheri-berry/%s/%s", request.Name, filename), reader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "package uploaded"})
}

func GetPackageInfo(c *gin.Context) {
	packageName := c.Param("package")
	packageJsonFile := fmt.Sprintf("cheri-berry/%s/package.json", packageName)

	exists, err := storage.Storage.Exists(context.TODO(), packageJsonFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package does not exist"})
		return
	}

	content, err := storage.Storage.GetBytes(context.TODO(), packageJsonFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pkgJson PackageJson
	err = json.Unmarshal(content, &pkgJson)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pkgJson)
}

func GetPackageFile(c *gin.Context) {
	packageName := c.Param("package")
	fileName := c.Param("file")

	filePath := fmt.Sprintf("cheri-berry/%s/%s", packageName, fileName)

	exists, err := storage.Storage.Exists(context.TODO(), filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package does not exist"})
		return
	}

	content, err := storage.Storage.GetBytes(context.TODO(), filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func convertPkgJson(req requests.UploadPackageRequest, packageJson *PackageJson) PackageJson {
	var pkgJson PackageJson
	if packageJson == nil {
		pkgJson = PackageJson{
			Name:     req.Name,
			Versions: make(map[string]PackageVersion),
			DistTags: struct {
				Latest string `json:"latest"`
			}{
				Latest: req.DistTags.Latest,
			},
			Id: req.ID,
			Attachments: map[string]struct {
				Shasum string `json:"shasum"`
			}{},
		}
	} else {
		pkgJson = *packageJson
	}

	for version, uploadVersion := range req.Versions {
		pkgVersion := PackageVersion{
			Name:        uploadVersion.Name,
			Version:     uploadVersion.Version,
			Description: uploadVersion.Description,
			Main:        uploadVersion.Main,
			Scripts: struct {
				Test string `json:"test"`
			}{
				Test: uploadVersion.Scripts["test"],
			},
			Author: struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}{
				Name: uploadVersion.Author, // Assuming Author is just a name in UploadPackageRequest
			},
			License:     uploadVersion.License,
			Id:          uploadVersion.ID,
			NodeVersion: uploadVersion.NodeVersion,
			NpmVersion:  uploadVersion.NpmVersion,
			Dist: struct {
				Integrity string `json:"integrity"`
				Shasum    string `json:"shasum"`
				Tarball   string `json:"tarball"`
			}{
				Integrity: uploadVersion.Dist.Integrity,
				Shasum:    uploadVersion.Dist.Shasum,
				Tarball:   uploadVersion.Dist.Tarball,
			},
		}
		pkgJson.Versions[version] = pkgVersion
	}

	for key, attachment := range req.Attachments {
		pkgJson.Attachments[key] = struct {
			Shasum string `json:"shasum"`
		}{
			Shasum: attachment.Data, // Assuming the data field contains the shasum
		}
	}

	return pkgJson
}
