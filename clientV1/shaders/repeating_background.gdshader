shader_type canvas_item;

uniform sampler2D background_texture : source_color, repeat_enable;
uniform vec2 base_tile_size = vec2(960.0, 540.0);
uniform float tile_scale = 1.0;
uniform vec2 scroll_offset = vec2(0.0, 0.0);
uniform float opacity = 1.0;

void fragment() {
	vec2 pixel_position = FRAGCOORD.xy + scroll_offset;
	vec2 repeated_uv = fract(pixel_position / (base_tile_size * tile_scale));
	vec4 texture_color = texture(background_texture, repeated_uv);
	texture_color.a *= opacity;
	COLOR = texture_color;
}
