package filestorageimpl

import (
	"context"
	"image"
	"io"
	"mime/multipart"
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util/xerr"
)

// CloudflareR2 Cloudflare R2 存储实现
// 复用 MinIO 客户端，因为 R2 兼容 S3 API
type CloudflareR2 struct {
	OptionService service.OptionService
}

func NewCloudflareR2(optionService service.OptionService) *CloudflareR2 {
	return &CloudflareR2{
		OptionService: optionService,
	}
}

func (c *CloudflareR2) Upload(ctx context.Context, fileHeader *multipart.FileHeader) (*dto.AttachmentDTO, error) {
	clientInstance, err := c.getR2Client(ctx)
	if err != nil {
		return nil, err
	}

	fd, err := newURLFileDescriptor(
		withBaseURL(clientInstance.Protocol+clientInstance.EndPoint+"/"+clientInstance.BucketName),
		withSubURLPath(clientInstance.Source),
		withShouldRenameURLOption(commonRenamePredicateFunc(ctx, consts.AttachmentTypeCloudflareR2)),
		withOriginalNameURLOption(fileHeader.Filename),
	)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("open upload file error")
	}
	defer file.Close()
	_, err = clientInstance.PutObject(ctx, clientInstance.BucketName, fd.getRelativePath(), file, fileHeader.Size, minio.PutObjectOptions{})
	if err != nil {
		return nil, xerr.WithMsg(err, "upload to R2 error").WithStatus(xerr.StatusInternalServerError).WithErrMsgf("err=%v", err)
	}

	mediaType, _ := getFileContentType(file)
	result := &dto.AttachmentDTO{
		Name:           fd.getFileName(),
		Path:           fd.getRelativePath(),
		FileKey:        fd.getRelativePath(),
		Suffix:         fd.getExtensionName(),
		MediaType:      mediaType,
		AttachmentType: consts.AttachmentTypeCloudflareR2,
		Size:           fileHeader.Size,
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError)
	}
	err = handleImageMeta(file, result, func(srcImage image.Image) (string, error) {
		return fd.getRelativePath(), nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CloudflareR2) Delete(ctx context.Context, fileKey string) error {
	clientInstance, err := c.getR2Client(ctx)
	if err != nil {
		return err
	}
	err = clientInstance.RemoveObject(ctx, clientInstance.BucketName, fileKey, minio.RemoveObjectOptions{})
	if err != nil {
		return xerr.WithStatus(err, xerr.StatusInternalServerError).WithErrMsgf("err=%v", err)
	}
	return nil
}

func (c *CloudflareR2) GetAttachmentType() consts.AttachmentType {
	return consts.AttachmentTypeCloudflareR2
}

func (c *CloudflareR2) GetFilePath(ctx context.Context, relativePath string) (string, error) {
	clientInstance, err := c.getR2Client(ctx)
	if err != nil {
		return "", err
	}
	base := clientInstance.Protocol + clientInstance.EndPoint + "/" + clientInstance.BucketName
	if clientInstance.FrontBase != "" {
		base = clientInstance.FrontBase
	}
	fullPath, _ := url.JoinPath(base, relativePath)
	fullPath, _ = url.PathUnescape(fullPath)
	return fullPath, nil
}

type r2Client struct {
	*minio.Client
	BucketName string
	Source     string
	EndPoint   string
	Protocol   string
	FrontBase  string
}

func (c *CloudflareR2) getR2Client(ctx context.Context) (*r2Client, error) {
	getClientProperty := func(propertyValue *string, property property.Property, allowEmpty bool, e error) error {
		if e != nil {
			return e
		}
		value, err := c.OptionService.GetOrByDefaultWithErr(ctx, property, property.DefaultValue)
		if err != nil {
			return err
		}
		strValue, ok := value.(string)
		if !ok {
			return xerr.WithStatus(nil, xerr.StatusBadRequest).WithErrMsgf("wrong property type")
		}
		if !allowEmpty && strValue == "" {
			return xerr.WithStatus(nil, xerr.StatusInternalServerError).WithMsg("property not found: " + property.KeyValue)
		}
		*propertyValue = strValue
		return nil
	}
	// 复用 MinIO 的配置项
	var endPoint, bucketName, accessKey, accessSecret, protocol, source, region, frontBase string
	err := getClientProperty(&endPoint, property.MinioEndpoint, false, nil)
	err = getClientProperty(&bucketName, property.MinioBucketName, false, err)
	err = getClientProperty(&accessKey, property.MinioAccessKey, false, err)
	err = getClientProperty(&accessSecret, property.MinioAccessSecret, false, err)
	err = getClientProperty(&protocol, property.MinioProtocol, false, err)
	err = getClientProperty(&source, property.MinioSource, true, err)
	err = getClientProperty(&region, property.MinioRegion, true, err)
	err = getClientProperty(&frontBase, property.MinioFrontBase, true, err)
	if err != nil {
		return nil, err
	}
	secure := func() bool {
		switch protocol {
		case "https://":
			return true
		case "http://":
			return false
		default:
			return true
		}
	}()
	client, err := minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, accessSecret, ""),
		Secure: secure,
		Region: region,
	})
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("failed to initialize R2: " + err.Error())
	}

	clientInstance := &r2Client{}
	clientInstance.Client = client
	clientInstance.BucketName = bucketName
	clientInstance.Source = source
	clientInstance.EndPoint = endPoint
	clientInstance.Protocol = protocol
	clientInstance.FrontBase = frontBase
	return clientInstance, nil
}
