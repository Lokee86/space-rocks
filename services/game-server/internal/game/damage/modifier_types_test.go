package damage

import "testing"

func TestDamageModifierCategoryStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "outgoing", got: string(DamageModifierCategoryOutgoing), want: "outgoing"},
		{name: "resistance", got: string(DamageModifierCategoryResistance), want: "resistance"},
		{name: "vulnerability", got: string(DamageModifierCategoryVulnerability), want: "vulnerability"},
		{name: "generic", got: string(DamageModifierCategoryGeneric), want: "generic"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}

func TestDamageModifierOperationStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "add", got: string(DamageModifierOperationAdd), want: "add"},
		{name: "multiply", got: string(DamageModifierOperationMultiply), want: "multiply"},
		{name: "set", got: string(DamageModifierOperationSet), want: "set"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}

func TestDamageCauseStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "collision", got: string(DamageCauseCollision), want: "collision"},
		{name: "projectile", got: string(DamageCauseProjectile), want: "projectile"},
		{name: "debug", got: string(DamageCauseDebug), want: "debug"},
		{name: "area", got: string(DamageCauseArea), want: "area"},
		{name: "dot", got: string(DamageCauseDot), want: "dot"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}

func TestDamageTypeStringValues(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "kinetic", got: string(DamageTypeKinetic), want: "kinetic"},
		{name: "explosive", got: string(DamageTypeExplosive), want: "explosive"},
		{name: "energy", got: string(DamageTypeEnergy), want: "energy"},
		{name: "thermal", got: string(DamageTypeThermal), want: "thermal"},
		{name: "radioactive", got: string(DamageTypeRadioactive), want: "radioactive"},
		{name: "true_damage", got: string(DamageTypeTrueDamage), want: "true_damage"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, tc.got)
			}
		})
	}
}
