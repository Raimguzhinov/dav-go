package grpc

import (
	"context"

	caldavGRPC "github.com/Raimguzhinov/dav-go/internal/delivery/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	caldavGRPC.UnimplementedCalendarServer
}

func New() caldavGRPC.CalendarServer {
	return &grpcServer{}
}

func (s *grpcServer) FolderList(context.Context, *caldavGRPC.FolderListRequest) (*caldavGRPC.FolderListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FolderList not implemented")
}

func (s *grpcServer) GetFolder(context.Context, *caldavGRPC.FolderRequest) (*caldavGRPC.FolderInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFolder not implemented")
}

func (s *grpcServer) CreateFolder(context.Context, *caldavGRPC.CreateFolderRequest) (*caldavGRPC.FolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateFolder not implemented")
}

func (s *grpcServer) DeleteFolder(context.Context, *caldavGRPC.FolderRequest) (*caldavGRPC.FolderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFolder not implemented")
}

func (s *grpcServer) CalendarObjectList(context.Context, *caldavGRPC.FolderRequest) (*caldavGRPC.CalendarObjectListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CalendarObjectList not implemented")
}

func (s *grpcServer) GetCalendarObject(context.Context, *caldavGRPC.CalendarObjectRequest) (*caldavGRPC.CalendarObjectInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCalendarObject not implemented")
}

func (s *grpcServer) PutCalendarObject(context.Context, *caldavGRPC.CalendarObjectInfo) (*caldavGRPC.PutCalendarObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutCalendarObject not implemented")
}

func (s *grpcServer) DeleteEvent(context.Context, *caldavGRPC.CalendarObjectRequest) (*caldavGRPC.DeleteCalendarObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteEvent not implemented")
}
