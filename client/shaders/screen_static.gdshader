shader_type canvas_item;
render_mode blend_mix;

uniform float scanline_count = 480.0;
uniform float scanline_strength = 0.22;
uniform float scanline_hardness = 1.65;

// Keep this at 0.0 unless you specifically want screen-wide static.
uniform float grain_strength = 0.0;
uniform float grain_speed = 45.0;
uniform float grain_scale = 900.0;

uniform float flicker_strength = 0.045;
uniform float flicker_speed = 18.0;
uniform float flicker_speed_variance = 0.55;
uniform float flicker_variance_speed = 1.35;
uniform float flicker_secondary_strength = 0.35;

// Makes the scanlines themselves less solid.
uniform float scanline_breakup_strength = 0.16;
uniform float scanline_breakup_segments = 36.0;
uniform float scanline_breakup_cutoff = 0.46;
uniform float scanline_breakup_softness = 0.18;
uniform float scanline_line_variance_strength = 0.08;

uniform float roll_strength = 0.12;
uniform float roll_interval = 5.0;
uniform float roll_duration = 1.2;
uniform float roll_width = 0.10;

uniform float horizontal_shimmer_strength = 0.10;
uniform float horizontal_shimmer_speed = 1.8;
uniform float horizontal_shimmer_count = 42.0;

// Makes additive effects harsher / less hazy.
uniform float effect_cutoff = 0.018;
uniform float effect_gain = 1.45;

// Dogbone-style edge glow.
// Thin at edge centers, wider around corners.
uniform float edge_glow_strength = 0.12;
uniform float edge_glow_width = 0.018;
uniform float edge_corner_glow_width = 0.10;
uniform float edge_corner_glow_power = 2.2;
uniform float edge_glow_softness = 0.018;

// Normal simple vignette.
uniform float vignette_strength = 0.32;

// Flat edge bypass.
// Prevents the vignette from crushing the very edge,
// but does NOT make the vignette dogbone-shaped.
uniform float vignette_edge_bypass_strength = 1.0;
uniform float vignette_edge_bypass_width = 0.035;

float hash(vec2 p) {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

void fragment() {
	vec2 uv = UV;

	// Uses the ColorRect color.
	vec4 source = COLOR;
	vec3 color = source.rgb;

	// -------------------------
	// Scanlines
	// -------------------------
	float scanline_wave = sin(uv.y * scanline_count);
	float scanline = 0.5 + 0.5 * scanline_wave;

	// Sharper scanline shape. Higher scanline_hardness = less soft.
	scanline = pow(scanline, max(scanline_hardness, 0.001));

	// Uneven scanline-only flicker timing.
	float flicker_phase = TIME * flicker_speed;

	flicker_phase += sin(TIME * flicker_variance_speed) * flicker_speed_variance;
	flicker_phase += sin(TIME * flicker_variance_speed * 2.17 + 1.9) * flicker_speed_variance * 0.45;
	flicker_phase += sin(TIME * flicker_variance_speed * 0.41 + 4.2) * flicker_speed_variance * 0.65;

	float flicker_signal = sin(flicker_phase);
	flicker_signal += sin(flicker_phase * 2.31 + 1.3) * flicker_secondary_strength;
	flicker_signal += sin(flicker_phase * 4.17 + 3.8) * flicker_secondary_strength * 0.35;

	flicker_signal /= 1.0 + flicker_secondary_strength + flicker_secondary_strength * 0.35;

	float flicker = 1.0 + flicker_signal * flicker_strength;

	// One stable ID per scanline.
	float tau = 6.28318530718;
	float line_id = floor((uv.y * scanline_count) / tau);

	// Coarse horizontal breakup along each scanline.
	float segment_x = uv.x * scanline_breakup_segments;
	float segment_id = floor(segment_x);
	float segment_t = fract(segment_x);
	segment_t = segment_t * segment_t * (3.0 - 2.0 * segment_t);

	float break_a = hash(vec2(segment_id, line_id));
	float break_b = hash(vec2(segment_id + 1.0, line_id));
	float break_noise = mix(break_a, break_b, segment_t);

	float break_keep = smoothstep(
		scanline_breakup_cutoff - scanline_breakup_softness,
		scanline_breakup_cutoff + scanline_breakup_softness,
		break_noise
	);

	// Per-line strength variation.
	float line_variance = hash(vec2(line_id, 19.37));
	line_variance = mix(
		1.0 - scanline_line_variance_strength,
		1.0 + scanline_line_variance_strength,
		line_variance
	);

	// Reduce scanline darkness in broken segments.
	float line_breakup = mix(
		1.0,
		break_keep,
		scanline_breakup_strength
	);

	float scanline_darken = scanline * scanline_strength * flicker * line_breakup * line_variance;

	// -------------------------
	// Grain
	// -------------------------
	float grain = hash(floor(uv * grain_scale) + floor(TIME * grain_speed));
	grain = (grain - 0.5) * grain_strength;

	// -------------------------
	// Periodic roll
	// -------------------------
	float roll_time = mod(TIME, max(roll_interval, 0.001));
	float roll_active = 1.0 - step(roll_duration, roll_time);

	float roll_progress = roll_time / max(roll_duration, 0.001);
	roll_progress = clamp(roll_progress, 0.0, 1.0);

	float roll_fade_in = smoothstep(0.0, 0.12, roll_progress);
	float roll_fade_out = 1.0 - smoothstep(0.88, 1.0, roll_progress);
	float roll_gate = roll_active * roll_fade_in * roll_fade_out;

	float roll_dist = abs(uv.y - roll_progress);
	float roll = 1.0 - smoothstep(0.0, roll_width, roll_dist);
	roll *= roll_strength * roll_gate;

	// -------------------------
	// Horizontal shimmer
	// -------------------------
	float shimmer = sin((uv.y * horizontal_shimmer_count) + (TIME * horizontal_shimmer_speed));
	shimmer = 0.5 + 0.5 * shimmer;
	shimmer *= horizontal_shimmer_strength;

	// -------------------------
	// Dogbone edge glow
	// -------------------------
	float edge_dist = min(
		min(uv.x, 1.0 - uv.x),
		min(uv.y, 1.0 - uv.y)
	);

	// 0 near edge centers, 1 near corners.
	vec2 corner_uv = abs(uv * 2.0 - 1.0);
	float cornerness = corner_uv.x * corner_uv.y;
	cornerness = pow(cornerness, max(edge_corner_glow_power, 0.001));

	float dynamic_edge_width = edge_glow_width + edge_corner_glow_width * cornerness;

	float edge_glow = 1.0 - smoothstep(
		dynamic_edge_width,
		dynamic_edge_width + edge_glow_softness,
		edge_dist
	);

	edge_glow *= edge_glow_strength;

	// -------------------------
	// Normal vignette
	// -------------------------
	vec2 centered = uv * 2.0 - 1.0;
	float vignette = 1.0 - dot(centered, centered) * vignette_strength;
	vignette = clamp(vignette, 0.0, 1.0);

	// -------------------------
	// Flat edge vignette bypass  
	// -------------------------
	float bypass_left = 1.0 - smoothstep(0.0, vignette_edge_bypass_width, uv.x);
	float bypass_right = 1.0 - smoothstep(0.0, vignette_edge_bypass_width, 1.0 - uv.x);
	float bypass_top = 1.0 - smoothstep(0.0, vignette_edge_bypass_width, uv.y);
	float bypass_bottom = 1.0 - smoothstep(0.0, vignette_edge_bypass_width, 1.0 - uv.y);

	float vignette_bypass = max(max(bypass_left, bypass_right), max(bypass_top, bypass_bottom));
	vignette_bypass *= vignette_edge_bypass_strength;
	vignette_bypass = clamp(vignette_bypass, 0.0, 1.0);

	float final_vignette = mix(vignette, 1.0, vignette_bypass);

	// -------------------------
	// Final composition
	// -------------------------
	float brighten = grain + roll + shimmer + edge_glow;

	// Harshness pass: cuts low haze and boosts remaining signal.
	brighten = max(brighten - effect_cutoff, 0.0) * effect_gain;

	color *= 1.0 - scanline_darken;
	color += source.rgb * brighten;

	// Vignette affects the screen body, but the flat edge bypass
	// keeps the edge region from being crushed.
	color *= final_vignette;

	COLOR = vec4(color, source.a);
}