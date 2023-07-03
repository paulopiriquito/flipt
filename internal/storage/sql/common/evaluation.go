package common

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}
	// get all rules for flag with their constraints if any
	rows, err := s.builder.Select("r.id, r.namespace_key, r.flag_key, r.segment_key, s.match_type, r.\"rank\", c.id, c.type, c.property, c.operator, c.value").
		From("rules r").
		Join("segments s ON (r.segment_key = s.\"key\" AND r.namespace_key = s.namespace_key)").
		LeftJoin("constraints c ON (s.\"key\" = c.segment_key AND s.namespace_key = c.namespace_key)").
		Where(sq.And{sq.Eq{"r.flag_key": flagKey}, sq.Eq{"r.namespace_key": namespaceKey}}).
		OrderBy("r.\"rank\" ASC").
		GroupBy("r.id, c.id, s.match_type").
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var (
		uniqueRules = make(map[string]*storage.EvaluationRule)
		rules       = []*storage.EvaluationRule{}
	)

	for rows.Next() {
		var (
			tempRule           storage.EvaluationRule
			optionalConstraint optionalConstraint
		)

		if err := rows.Scan(
			&tempRule.ID,
			&tempRule.NamespaceKey,
			&tempRule.FlagKey,
			&tempRule.SegmentKey,
			&tempRule.SegmentMatchType,
			&tempRule.Rank,
			&optionalConstraint.Id,
			&optionalConstraint.Type,
			&optionalConstraint.Property,
			&optionalConstraint.Operator,
			&optionalConstraint.Value); err != nil {
			return rules, err
		}

		if existingRule, ok := uniqueRules[tempRule.ID]; ok {
			// current rule we know about
			if optionalConstraint.Id.Valid {
				constraint := storage.EvaluationConstraint{
					ID:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
				existingRule.Constraints = append(existingRule.Constraints, constraint)
			}
		} else {
			// haven't seen this rule before
			newRule := &storage.EvaluationRule{
				ID:               tempRule.ID,
				NamespaceKey:     tempRule.NamespaceKey,
				FlagKey:          tempRule.FlagKey,
				SegmentKey:       tempRule.SegmentKey,
				SegmentMatchType: tempRule.SegmentMatchType,
				Rank:             tempRule.Rank,
			}

			if optionalConstraint.Id.Valid {
				constraint := storage.EvaluationConstraint{
					ID:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
				newRule.Constraints = append(newRule.Constraints, constraint)
			}

			uniqueRules[newRule.ID] = newRule
			rules = append(rules, newRule)
		}
	}

	if err := rows.Err(); err != nil {
		return rules, err
	}

	if err := rows.Close(); err != nil {
		return rules, err
	}

	return rules, nil
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	rows, err := s.builder.Select("d.id, d.rule_id, d.variant_id, d.rollout, v.\"key\", v.attachment").
		From("distributions d").
		Join("variants v ON (d.variant_id = v.id)").
		Where(sq.Eq{"d.rule_id": ruleID}).
		OrderBy("d.created_at ASC").
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var distributions []*storage.EvaluationDistribution

	for rows.Next() {
		var (
			d          storage.EvaluationDistribution
			attachment sql.NullString
		)

		if err := rows.Scan(
			&d.ID, &d.RuleID, &d.VariantID, &d.Rollout, &d.VariantKey, &attachment,
		); err != nil {
			return distributions, err
		}

		if attachment.Valid {
			attachmentString, err := compactJSONString(attachment.String)
			if err != nil {
				return distributions, err
			}
			d.VariantAttachment = attachmentString
		}

		distributions = append(distributions, &d)
	}

	if err := rows.Err(); err != nil {
		return distributions, err
	}

	if err := rows.Close(); err != nil {
		return distributions, err
	}

	return distributions, nil
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRollout, error) {
	return []*storage.EvaluationRollout{}, nil
}
