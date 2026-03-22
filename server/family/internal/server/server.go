package server

import (
	"context"
	"strconv"

	familyv1 "github.com/boykush/famoney/server/family/gen/go"
	"github.com/boykush/famoney/server/family/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	familyv1.UnimplementedFamilyServiceServer
	memberUsecase *usecase.MemberUsecase
}

func NewServer(memberUsecase *usecase.MemberUsecase) *Server {
	return &Server{memberUsecase: memberUsecase}
}

func (s *Server) HealthCheck(ctx context.Context, req *familyv1.HealthCheckRequest) (*familyv1.HealthCheckResponse, error) {
	return &familyv1.HealthCheckResponse{
		Status: familyv1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *Server) GetMemberBySubject(ctx context.Context, req *familyv1.GetMemberBySubjectRequest) (*familyv1.Member, error) {
	m, err := s.memberUsecase.GetByOIDCSubject(ctx, req.GetOidcSubject())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "member not found: %v", err)
	}
	return &familyv1.Member{
		Id:          strconv.Itoa(m.ID),
		OidcSubject: m.OIDCSubject,
		DisplayName: m.DisplayName,
	}, nil
}

func (s *Server) CreateMember(ctx context.Context, req *familyv1.CreateMemberRequest) (*familyv1.Member, error) {
	m, err := s.memberUsecase.Create(ctx, req.GetOidcSubject(), req.GetDisplayName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create member: %v", err)
	}
	return &familyv1.Member{
		Id:          strconv.Itoa(m.ID),
		OidcSubject: m.OIDCSubject,
		DisplayName: m.DisplayName,
	}, nil
}
