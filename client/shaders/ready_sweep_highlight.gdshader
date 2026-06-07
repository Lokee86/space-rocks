shader_type canvas_item;
render_mode blend_add;

uniform float sweep_progress : hint_range(-1.0, 2.0) = 0.0;
uniform float sweep_width : hint_range(0.001, 1.0) = 0.12;
uniform float sweep_softness : hint_range(0.001, 1.0) = 0.08;
uniform float sweep_alpha : hint_range(0.0, 2.0) = 0.75;
uniform float sweep_angle : hint_range(-3.14159, 3.14159) = 0.523599;
uniform float badge_radius : hint_range(0.0, 1.0) = 0.5;
uniform float badge_softness : hint_range(0.001, 1.0) = 0.02;


void fragment() {
	vec2 direction = vec2(cos(sweep_angle), sin(sweep_angle));
	vec2 centered_uv = UV - vec2(0.5);
	float projected = dot(centered_uv, direction) + 0.5;

	float half_width = max(sweep_width * 0.5, 0.0001);
	float softness = max(sweep_softness, 0.0001);

	float inner = 1.0 - smoothstep(0.0, half_width, abs(projected - sweep_progress));
	float outer = 1.0 - smoothstep(half_width, half_width + softness, abs(projected - sweep_progress));
	float band = clamp(max(inner, outer), 0.0, 1.0);
	float radial_distance = length(centered_uv);
	float mask = 1.0 - smoothstep(badge_radius, badge_radius + badge_softness, radial_distance);
	vec4 base_color = COLOR;
	float final_alpha = band * mask * sweep_alpha * base_color.a;

	COLOR = vec4(base_color.rgb * final_alpha, final_alpha);
}
