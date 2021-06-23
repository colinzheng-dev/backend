package server

// disabled SSO work for now due to missing encryption key in kubernetes manifest
// func (s *Server) createSSOSecret(w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	// Check whether the modification is allowed: user is administrator
// 	// or is organisation administrator for the organisation.
// 	org, allowed, err := s.orgModAllowed(w, r, false)
// 	if !allowed {
// 		return nil, err
// 	}

// 	rawSecret := chassis.NewBareID(32)

// 	encoded, err := chassis.Encrypt(rawSecret,s.encryptionKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := s.db.RotateSSOSecret(encoded, org.ID); err != nil {
// 		if err == db.ErrOrgNotFound {
// 			return chassis.NotFound(w)
// 		}
// 		return nil, err
// 	}

// 	response := messages.SSOSecret{
// 		Secret: &rawSecret,
// 	}

// 	return &response, nil
// }

// func (s *Server) dropSSOSecret(w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	// Check whether the modification is allowed: user is administrator
// 	// or is organisation administrator for the organisation.
// 	org, allowed, err := s.orgModAllowed(w, r, false)
// 	if !allowed {
// 		return nil, err
// 	}

// 	if err := s.db.DeleteSSOSecret(org.ID); err != nil {
// 		if err == db.ErrOrgNotFound {
// 			return chassis.NotFound(w)
// 		}
// 		return nil, err
// 	}

// 	return chassis.NoContent(w)
// }

// func (s *Server) getSSOSecret(w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	// Check whether the modification is allowed: user is administrator
// 	// or is organisation administrator for the organisation.
// 	org, allowed, err := s.orgModAllowed(w, r, false)
// 	if !allowed {
// 		return nil, err
// 	}

// 	secret, err := s.db.GetSSOSecretByOrgIDOrSlug(org.ID)
// 	if err != nil {
// 		if err == db.ErrOrgNotFound {
// 			return chassis.NotFound(w)
// 		}
// 		return nil, err
// 	}

// 	if secret.Secret != nil {
// 		return chassis.Decrypt(*secret.Secret, s.encryptionKey)
// 	}

// 	return chassis.NotFoundWithMessage(w, "sso secret is not set")
// }

// func (s *Server) getSSOSecretInternal(w http.ResponseWriter, r *http.Request) (interface{}, error) {
// 	idOrSlug := chi.URLParam(r, "id_or_slug")

// 	secret, err := s.db.GetSSOSecretByOrgIDOrSlug(idOrSlug)
// 	if err != nil {
// 		if err == db.ErrOrgNotFound {
// 			return chassis.NotFound(w)
// 		}
// 		return nil, err
// 	}

// 	if secret.Secret != nil {
// 		return chassis.Decrypt(*secret.Secret, s.encryptionKey)
// 	}

// 	return chassis.NotFoundWithMessage(w, "sso secret is not set")
// }
