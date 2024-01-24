package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poke-factory/cheri-berry/internal/requests"
	"github.com/poke-factory/cheri-berry/internal/storage"
	"golang.org/x/net/context"
	"net/http"
	"strings"
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
		pkgJson, err = convertPkgJson(request, &pkgJson)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		pkgJson, _ = convertPkgJson(request, nil)
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

func DeletePackageInfo(c *gin.Context) {
	packageName := c.Param("package")
	packageJsonFile := fmt.Sprintf("cheri-berry/%s/package.json", packageName)
	exits, err := storage.Storage.Exists(context.Background(), packageJsonFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !exits {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package does not exist"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": "found package"})
}

func DeletePackage(c *gin.Context) {
	packageName := c.Param("package")
	files, err := storage.Storage.Files(context.TODO(), fmt.Sprintf("cheri-berry/%s", packageName))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, file := range files {
		err = storage.Storage.Delete(context.Background(), file.Key())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "package deleted"})
}

func DeletePackageVersion(c *gin.Context) {
	packageName := c.Param("package")
	fileName := c.Param("file")

	err := storage.Storage.Delete(context.TODO(), fmt.Sprintf("cheri-berry/%s/%s", packageName, fileName))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	packageJsonFile := fmt.Sprintf("cheri-berry/%s/package.json", packageName)

	f, err := storage.Storage.GetBytes(context.Background(), packageJsonFile)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pkgJson PackageJson
	err = json.Unmarshal(f, &pkgJson)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delVersion := strings.Replace(fileName, ".tgz", "", 1)
	delVersion = strings.Replace(delVersion, pkgJson.Name+"-", "", 1)

	delete(pkgJson.Versions, delVersion)
	delete(pkgJson.Time, delVersion)
	delete(pkgJson.Attachments, fileName)

	if delVersion == pkgJson.DistTags.Latest {
		pkgJson.DistTags.Latest = ""
		for version := range pkgJson.Versions {
			if pkgJson.DistTags.Latest == "" {
				pkgJson.DistTags.Latest = version
			} else {
				v1, _ := semver.NewVersion(pkgJson.DistTags.Latest)
				v2, _ := semver.NewVersion(version)
				if v2.GreaterThan(v1) {
					pkgJson.DistTags.Latest = version
				}
			}
		}
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
	c.JSON(http.StatusOK, gin.H{"message": "version deleted"})
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

func convertPkgJson(req requests.UploadPackageRequest, packageJson *PackageJson) (PackageJson, error) {
	var pkgJson PackageJson
	if packageJson == nil {
		pkgJson = PackageJson{
			Rev:      uuid.New().String(),
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
			Time: map[string]time.Time{
				"created": time.Now(),
			},
		}
	} else {
		pkgJson = *packageJson
		v1, err := semver.NewVersion(req.Versions[req.DistTags.Latest].Version)
		if err != nil {
			return pkgJson, err
		}
		v2, err := semver.NewVersion(pkgJson.Versions[pkgJson.DistTags.Latest].Version)
		if err != nil {
			return pkgJson, err
		}
		if v1.GreaterThan(v2) {
			pkgJson.DistTags.Latest = req.DistTags.Latest
		}
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
			}{},
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
		pkgJson.Time["modified"] = time.Now()
		pkgJson.Time[version] = time.Now()
	}

	for key, attachment := range req.Attachments {
		pkgJson.Attachments[key] = struct {
			Shasum string `json:"shasum"`
		}{
			Shasum: attachment.Data, // Assuming the data field contains the shasum
		}
	}
	return pkgJson, nil
}
