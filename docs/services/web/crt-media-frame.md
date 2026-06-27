# CRT Media Frame

Parent index: [Web](./!INDEX.md)

## Purpose

This document describes the current CRT-style media frame implementation used by the Space Rocks web site.

It defines the component’s runtime responsibilities, media-mode behavior, viewport framing model, control behavior, code ownership, and known implementation boundaries.

## Overview

`CrtMediaFrame` is the shared web presentation component for framed media on the Astro devlog site. It provides a CRT-styled shell that can present image galleries, YouTube videos, native video sources, or fallback child content inside a calibrated viewport.

The component owns the reusable frame shell and playback controls. It does not own devlog content, media asset storage, route generation, or Plasmic layout source.

The current public component is intentionally shared across image and video use cases. Image/video behavior should remain internally separated without creating separate public frame components unless the shared component becomes too fragile to maintain.

Current media-frame behavior includes:

```text
image galleries
-> Previous / Play-Pause / Next controls
-> auto-scroll by default when multiple images exist

YouTube and native video
-> RW / Play-Pause / FF controls

CRT shell
-> frame art
-> calibrated viewport insets
-> shader canvas overlay
-> bottom tray controls
```

The component is used by Plasmic-generated layout instances, but custom behavior belongs in owned source files, not generated Plasmic output.

## Code root

The current implementation lives under:

```text
web-astro/src/components/media/
```

Primary component:

```text
web-astro/src/components/media/CrtMediaFrame.tsx
```

Supporting files:

```text
web-astro/src/components/media/CrtMediaFrame.module.css
web-astro/src/components/media/CrtShaderCanvas.tsx
web-astro/src/components/media/MediaControlButton.tsx
web-astro/src/components/media/MediaControlButton.module.css
web-astro/src/components/media/index.ts
```

Known consumers include:

```text
web-astro/src/components/Homepage.tsx
web-astro/src/components/plasmic/space_rocks_devlog/PlasmicHomepage.tsx
```

`Homepage.tsx` is an owned wrapper. Generated Plasmic files may instantiate or configure the component, but they must not become the home for custom behavior.

## Responsibilities

`CrtMediaFrame` owns:

```text
CRT media shell rendering
viewport positioning inside the shell
media layer selection
image gallery rendering
image gallery auto-scroll state
native video rendering
YouTube iframe rendering
YouTube iframe API loading
media control mapping
CRT shader overlay placement
caption rendering
frame-art and control-tray composition
```

The component accepts content and configuration through props. It does not fetch devlog content or discover media assets itself.

The component supports these media inputs:

```text
imageItems
-> image gallery source list

src
-> fallback single image source

youtubeUrl
-> YouTube watch, short, or embed URL converted to iframe embed URL

videoSrc
-> native video source

children
-> fallback custom content
```

## Does not own

`CrtMediaFrame` does not own:

```text
devlog post content
devlog route generation
homepage/archive page composition
media asset folder conventions
Plasmic visual layout ownership
generated Plasmic source
YouTube video hosting
static-site deployment
analytics
gameplay runtime behavior
account/platform state
```

Devlog content and media paths belong to the static devlog site implementation. Plasmic owns visual layout source where practical. Owned wrappers connect content data to the frame component.

## Domain roles

Within the current website domain, `CrtMediaFrame` acts as a presentation component.

Role split:

```text
Devlog content
-> selects media kind and source values

Homepage wrapper
-> maps content fields into CrtMediaFrame props

Plasmic layout
-> places the frame in the visual page structure

CrtMediaFrame
-> renders the frame shell, media viewport, shader overlay, and controls

Static host
-> serves generated HTML, JS, CSS, and public media assets
```

The frame participates in the public website presentation flow only. It is not authoritative for content state or product/platform state.

## Protocols and APIs

`CrtMediaFrame` does not expose an external application API. It exposes a React component prop surface consumed by owned web components and Plasmic-generated layout instances.

The runtime surfaces it consumes are:

```text
browser media APIs
-> native <video> playback, pause, currentTime, duration

YouTube iframe API
-> player creation, play, pause, seek, playback state

static HTTP assets
-> image, video, frame-art, and UI asset delivery
```

YouTube URLs are converted into iframe embed URLs by `getYouTubeEmbedUrl()`.

Supported YouTube input forms include:

```text
https://www.youtube.com/watch?v=<id>
https://youtu.be/<id>
https://www.youtube.com/embed/<id>
```

Embed URLs include:

```text
enablejsapi=1
origin=<current browser origin>
```

The YouTube iframe API does not require a project API key for the current embed/control use case.

## Data ownership

`CrtMediaFrame` owns transient presentation state only.

Owned state includes:

```text
current image index
image slideshow playing/paused state
native video playing/paused state
YouTube player readiness state
YouTube playing/paused state
iframe/player refs
video refs
```

The component does not persist data.

Content values are provided from outside the component:

```text
image item paths
YouTube URL
video source
alt text
caption
media mode
control disabled state
viewport inset values
shader tuning values
```

Per-entry media storage is owned by the devlog static-site implementation, not by this component.

## Media modes

The component resolves media mode from available props.

YouTube takes precedence when `youtubeUrl` can be converted into a valid embed URL.

Resolution order:

```text
valid youtubeUrl
-> YouTube mode

mediaMode="video" with videoSrc
-> native video mode

mediaMode="imageList" or imageItems present
-> image gallery mode

src
-> fallback image mode

children
-> custom fallback content

no media
-> empty media viewport
```

The computed `data-media-kind` values are:

```text
youtube
images
video
empty
```

## Viewport and frame model

`aspectRatio` now means the desired visible viewport ratio, not the outer shell ratio.

The component computes the required outer shell ratio from the viewport target and the screen insets:

```text
shellAspectRatio = viewportAspectRatio * (1 - topInset - bottomInset) / (1 - leftInset - rightInset)
```

The default visible viewport target is:

```text
16 / 9
```

The media frame width should come from the parent layout, while the height should come from the component's computed aspect ratio.

The media viewport is absolutely positioned inside the shell using inset props:

```text
screenInsetLeft
screenInsetRight
screenInsetTop
screenInsetBottom
```

Numeric inset values are formatted as percentages.

The component default inset values are:

```text
screenInsetLeft = 5%
screenInsetRight = 5%
screenInsetTop = 5%
screenInsetBottom = 10%
```

The current visually calibrated page-level target is:

```text
left: 5%
right: 5%
top: 11%
bottom: 15%
```

Those values define the visible media screen inside the CRT frame art. They are visual calibration values, not normal page margins.

Plasmic instances should not hardcode `aspectRatio` just to make the shell look right. `aspectRatio` is the visible viewport contract, so shell sizing should follow from the viewport target plus the insets.

The frame art and bottom tray are separate from the viewport. The frame uses `border-image` sourced from:
```text
/assets/ui/media_frame.png
```

Frame dimensions use container query units so the border and tray scale with the frame shell.

## Layer order

The intended render layer order is:

```text
root figure
  shell
    viewport
      mediaLayer
        image / iframe / video / children
      shader canvas
    frame art
    bottom tray slot
      controls
  caption
```

The intended visual stacking order is:

```text
viewport background
media layer
CRT shader overlay
frame art
bottom tray / controls
```

Implementation notes:

```text
viewport
-> absolute inset screen area
-> overflow hidden
-> isolated paint area
-> dark CRT background

mediaLayer
-> absolute full viewport layer
-> contains actual media

shaderCanvas
-> absolute full viewport overlay
-> pointer-events none

frame
-> absolute full shell border art
-> pointer-events none

bottomTraySlot
-> absolute bottom tray area
-> controls can receive pointer events
```

Do not use negative z-index values in the media frame. The viewport and shell must preserve a stable positioning context.

## Image gallery behavior

Image galleries are driven by `imageItems`.

`imageItems` may be:

```text
string[]
newline-separated string
comma-separated string
```

The component trims and filters empty entries before rendering.

Image gallery behavior:

```text
multiple images
-> auto-scroll starts enabled

single image
-> no meaningful auto-scroll

Previous
-> moves to previous image and wraps around

Play/Pause
-> pauses or resumes timed auto-scroll

Next
-> moves to next image and wraps around
```

Default auto-scroll timing:

```text
autoAdvanceMs = 5000
```

When the image list changes, the component resets the index and restarts auto-scroll only when the new list has more than one image.

## Image sizing and asset preparation

The current practical image strategy is to prepare media assets for the calibrated CRT viewport instead of relying on further image-fit churn.

The component supports CSS object-fit behavior through the viewport `data-fit` value and image-specific styling, but image assets should still be composed for the visible frame shape.

For the current calibrated viewport:

```text
left + right inset = 10%
top + bottom inset = 26%
```

With a `16 / 9` visible viewport target and the current insets, the outer shell is wider than the viewport and is sized by the computed shell ratio. Avoid using an old shell-ratio shortcut such as `16 / 10.95`; that was the stale contract.

Plasmic Studio can render stale behavior if its parked host component copy differs from `web-astro`, so verify the live host wiring when the studio preview and runtime disagree.

Useful practical image sizes for frame-filling media are approximately:

```text
1920 x 890
1920 x 900
1600 x 740
1280 x 590
```

Exact crop and safe-area decisions should be verified visually against the CRT frame.

Important content should not be placed against the extreme outer edges of the image because the frame art and viewport calibration can obscure or compress the perceived screen area.

## YouTube behavior

When a valid `youtubeUrl` is supplied, the component renders a YouTube iframe and creates a YouTube player through the iframe API.

YouTube controls map to:

```text
left button
-> rewind by seekSeconds

center button
-> play or pause

right button
-> fast-forward by seekSeconds
```

Default seek interval:

```text
seekSeconds = 10
```

YouTube controls are enabled only after the iframe API reports the player as ready.

The YouTube lifecycle is sensitive. Avoid casual changes to:

```text
loadYouTubeIframeApi()
youtubeIframeRef
youtubePlayerRef
isYouTubeReady
isYouTubePlaying
player cleanup
iframe render path
```

If the iframe is visible but controls are disabled, suspect iframe API readiness first.

If the frame is blank, suspect the iframe render path, player lifecycle, or cleanup behavior before changing CSS.

## Native video behavior

When `mediaMode="video"` and `videoSrc` are present, the component renders a native `<video>` element.

Native video controls map to:

```text
left button
-> rewind by seekSeconds

center button
-> play or pause

right button
-> fast-forward by seekSeconds
```

The component tracks native video play state through `onPlay`, `onPause`, and `onEnded`.

The native video element uses:

```text
playsInline
preload="metadata"
```

Browser autoplay behavior is not owned by this component. Playback starts from user interaction through the media controls.

## Control behavior

The control tray always renders three logical button slots when controls are shown:

```text
previous
play
next
```

The visual button variant depends on media mode.

Image mode:

```text
previous
play / pause
next
```

YouTube and native video mode:

```text
rewind
play / pause
fastForward
```

Disabled controls can be provided through `disabledControls`.

`disabledControls` may be:

```text
string[]
comma-separated string
```

The component accepts both slot names and button variants when disabling controls.

Examples:

```text
previous
rewind
play
pause
fastForward
next
```

The play and pause variants are treated as related. Disabling one can disable the play/pause center action depending on the current state.

Controls should remain visible when media is present. Missing controls look broken; disabled controls should look intentional.

## Shader behavior

The CRT shader overlay is rendered by `CrtShaderCanvas`.

The shader is controlled through props passed from `CrtMediaFrame` to `CrtShaderCanvas`, including scanline, roll, shimmer, flicker, vignette, glow, and wave parameters.

The shader canvas is a visual overlay and must not intercept pointer events.

`shaderEnabled` controls whether the shader effect runs. The final enabled state also depends on the frame-level `enabled` prop.

## Code map

Primary implementation:

```text
web-astro/src/components/media/CrtMediaFrame.tsx
-> React component, media mode resolution, slideshow state, YouTube/native video control logic, viewport CSS variables, frame composition

web-astro/src/components/media/CrtMediaFrame.module.css
-> frame layout, viewport positioning, media layer styles, frame art, tray positioning, z-index order

web-astro/src/components/media/CrtShaderCanvas.tsx
-> CRT shader canvas rendering and shader animation parameters

web-astro/src/components/media/MediaControlButton.tsx
-> reusable visual media-control button component

web-astro/src/components/media/MediaControlButton.module.css
-> media-control button styling

web-astro/src/components/media/index.ts
-> media component exports
```

Current owned integration:

```text
web-astro/src/components/Homepage.tsx
-> maps content media fields into CrtMediaFrame props

web-astro/src/content/homepageContent.ts
-> normalizes homepage/devlog content fields consumed by Homepage

web-astro/src/content/devlog/
-> devlog content files that provide media values through the content pipeline
```

Generated integration:

```text
web-astro/src/components/plasmic/space_rocks_devlog/PlasmicHomepage.tsx
-> generated layout code that instantiates CrtMediaFrame

web-astro/src/components/plasmic/space_rocks_devlog/PlasmicHomepage.module.css
-> generated layout styling
```

Generated Plasmic files are not customization surfaces. Custom behavior must go in owned wrapper/source files.

Static assets:

```text
web-astro/public/assets/ui/media_frame.png
-> CRT frame art

web-astro/public/media/devlog/<slug>/
-> per-entry devlog media assets
```

## Tests

No dedicated automated test for `CrtMediaFrame` is currently documented.

Current verification is visual/runtime smoke testing through the Astro dev site.

Minimum smoke checks:

```text
devlog page renders
YouTube media appears in the CRT viewport
YouTube play/pause works
YouTube rewind/fast-forward works
image gallery appears in the CRT viewport
image gallery auto-scrolls by default
image previous/next controls work
image play/pause pauses and resumes auto-scroll
shader overlay appears above media
controls remain clickable
frame art remains aligned with the viewport
```

Static build verification belongs to the web static-site implementation, not this component alone.

## Related docs

* [Web](./!INDEX.md)
* [Devlog Static Site](devlog-static-site.md)
* [Plasmic / Astro Workflow](plasmic-astro-workflow.md)
* [Website and Web Presence](../../domains/web/website-and-web-presence.md)
* [Future Website and Web Presence Plan](../../planning/domains/web/website-and-web-presence.md)
* [Documentation Policy](../../documentation-policy.md)
* [Documentation Procedure](../../documentation-procedure.md)

## Notes

`CrtMediaFrame` is a fragile but currently working integration boundary. It combines frame geometry, media rendering, shader overlay, slideshow timing, YouTube iframe control, native video control, and control-button mapping.

Future cleanup should keep one public `CrtMediaFrame` shell but may split internal implementation branches:

```text
ImageMediaLayer
YoutubeMediaLayer
NativeVideoMediaLayer
useImageSlideshow()
useYoutubePlayer()
useNativeVideo()
```

The goal of that split would be to prevent image sizing changes from affecting YouTube behavior and prevent YouTube lifecycle changes from affecting image galleries.

Do not create separate public image/video frame components unless internal separation fails to keep the shared shell stable.
