shader_type canvas_item;

uniform sampler2D background_texture : source_color, repeat_enable;
uniform vec2 base_tile_world_size = vec2(960.0, 540.0);
uniform vec2 camera_world_position = vec2(0.0, 0.0);
uniform float camera_zoom = 1.0;
uniform vec2 viewport_stretch_scale = vec2(1.0, 1.0);
uniform float parallax_factor = 1.0;
uniform vec2 layer_world_offset = vec2(0.0, 0.0);
uniform float opacity = 1.0;

void fragment() {
	vec2 viewport_size = 1.0 / SCREEN_PIXEL_SIZE;
	vec2 screen_position_from_center = FRAGCOORD.xy - (viewport_size * 0.5);
	vec2 world_position = (
		(camera_world_position * parallax_factor)
		+ (screen_position_from_center / max(vec2(camera_zoom) * viewport_stretch_scale, vec2(0.001)))
		+ layer_world_offset
	);
	vec2 repeated_uv = fract(world_position / base_tile_world_size);
	vec4 texture_color = texture(background_texture, repeated_uv);
	texture_color.a *= opacity;
	COLOR = texture_color;
}
