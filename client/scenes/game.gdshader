shader_type canvas_item;

uniform vec2 tile_scale = vec2(1.0, 1.0);

void fragment() {
	vec2 tiled_uv = fract(UV * tile_scale);
	COLOR = texture(TEXTURE, tiled_uv);
}