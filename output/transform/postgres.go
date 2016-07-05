package transform

import (
	snapshot "github.com/pganalyze/collector/output/pganalyze_collector"
	"github.com/pganalyze/collector/state"
)

type OidToIdx map[state.Oid]int32

func transformPostgres(s snapshot.FullSnapshot, newState state.State, diffState state.DiffState) snapshot.FullSnapshot {
	s, roleOidToIdx := transformPostgresRoles(s, newState)
	s, databaseOidToIdx := transformPostgresDatabases(s, newState, roleOidToIdx)

	s = transformPostgresVersion(s, newState)
	s = transformPostgresConfig(s, newState)
	s = transformPostgresStatements(s, newState, diffState, roleOidToIdx, databaseOidToIdx)
	s = transformPostgresRelations(s, newState, diffState, roleOidToIdx, databaseOidToIdx)
	s = transformPostgresFunctions(s, newState, diffState, roleOidToIdx, databaseOidToIdx)

	return s
}

func transformPostgresRoles(s snapshot.FullSnapshot, newState state.State) (snapshot.FullSnapshot, OidToIdx) {
	roleOidToIdx := make(OidToIdx)

	for _, role := range newState.Roles {
		ref := snapshot.RoleReference{Name: role.Name}
		idx := int32(len(s.RoleReferences))
		s.RoleReferences = append(s.RoleReferences, &ref)
		roleOidToIdx[role.Oid] = idx
	}

	for _, role := range newState.Roles {
		info := snapshot.RoleInformation{
			RoleIdx:            roleOidToIdx[role.Oid],
			Inherit:            role.Inherit,
			Login:              role.Login,
			CreateDb:           role.CreateDb,
			CreateRole:         role.CreateRole,
			SuperUser:          role.SuperUser,
			Replication:        role.Replication,
			BypassRls:          role.BypassRLS,
			ConnectionLimit:    role.ConnectionLimit,
			PasswordValidUntil: snapshot.NullTimeToNullTimestamp(role.PasswordValidUntil),
			Config:             role.Config,
		}

		for _, oid := range role.MemberOf {
			info.MemberOf = append(info.MemberOf, roleOidToIdx[oid])
		}

		s.RoleInformations = append(s.RoleInformations, &info)
	}

	return s, roleOidToIdx
}

func transformPostgresDatabases(s snapshot.FullSnapshot, newState state.State, roleOidToIdx OidToIdx) (snapshot.FullSnapshot, OidToIdx) {
	databaseOidToIdx := make(OidToIdx)

	for _, database := range newState.Databases {
		ref := snapshot.DatabaseReference{Name: database.Name}
		idx := int32(len(s.DatabaseReferences))
		s.DatabaseReferences = append(s.DatabaseReferences, &ref)
		databaseOidToIdx[database.Oid] = idx
	}

	for _, database := range newState.Databases {
		info := snapshot.DatabaseInformation{
			DatabaseIdx:         databaseOidToIdx[database.Oid],
			OwnerRoleIdx:        roleOidToIdx[database.OwnerRoleOid],
			Encoding:            database.Encoding,
			Collate:             database.Collate,
			CType:               database.CType,
			IsTemplate:          database.IsTemplate,
			AllowConnections:    database.AllowConnections,
			ConnectionLimit:     database.ConnectionLimit,
			FrozenXid:           uint32(database.FrozenXID),
			MinimumMultixactXid: uint32(database.MinimumMultixactXID),
		}

		s.DatabaseInformations = append(s.DatabaseInformations, &info)
	}

	return s, databaseOidToIdx
}

func transformPostgresConfig(s snapshot.FullSnapshot, newState state.State) snapshot.FullSnapshot {
	for _, setting := range newState.Settings {
		info := snapshot.Setting{Name: setting.Name}

		if setting.CurrentValue.Valid {
			info.CurrentValue = setting.CurrentValue.String
		}
		if setting.Unit.Valid {
			info.Unit = &snapshot.NullString{Valid: true, Value: setting.Unit.String}
		}
		if setting.BootValue.Valid {
			info.BootValue = &snapshot.NullString{Valid: true, Value: setting.BootValue.String}
		}
		if setting.ResetValue.Valid {
			info.ResetValue = &snapshot.NullString{Valid: true, Value: setting.ResetValue.String}
		}
		if setting.Source.Valid {
			info.Source = &snapshot.NullString{Valid: true, Value: setting.Source.String}
		}
		if setting.SourceFile.Valid {
			info.SourceFile = &snapshot.NullString{Valid: true, Value: setting.SourceFile.String}
		}
		if setting.SourceLine.Valid {
			info.SourceLine = &snapshot.NullString{Valid: true, Value: setting.SourceLine.String}
		}

		s.Settings = append(s.Settings, &info)
	}

	return s
}

func transformPostgresVersion(s snapshot.FullSnapshot, newState state.State) snapshot.FullSnapshot {
	s.PostgresVersion = &snapshot.PostgresVersion{
		Full:    newState.Version.Full,
		Short:   newState.Version.Short,
		Numeric: int64(newState.Version.Numeric),
	}
	return s
}