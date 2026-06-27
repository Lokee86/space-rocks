---
title: "Testing Media Assets"
date: 2026-06-27
summary: "A media-focused devlog entry that exercises the YouTube player path and a full eight-image asteroid gallery."
heroLine1: "MEDIA ASSETS TEST"
heroLine2: "YOUTUBE PLUS ASTEROIDS"
heroLine3: "CRT FRAME CHECK"
articleLabel: "Devlog Entry 003"
articleTitle: "Testing Media Assets"
intro: "This entry exists to verify the real media path for the devlog shell. It uses one YouTube video for the headline media slot and a full set of asteroid images for the article media slot so we can confirm both playback modes in the live layout."
whatChanged: "The devlog content flow is now pointed at concrete media instead of placeholder assets. That means the same content-driven route can exercise the YouTube iframe path, the slideshow image path, and the existing CRT frame controls without requiring homepage-only test wiring."
callout: "If this entry renders correctly, the content collection can carry real media for both video and image gallery use cases."
whatsNext: "Next checks are simple polish passes: verifying the control labels feel right, confirming the archive order stays predictable, and making sure future entries can swap media without touching layout code."
finishedBody: "The reusable media frame can now host direct video, YouTube embeds, and image lists."
nowBody: "This entry is validating the public media asset pipeline with real asteroid art."
comingUpBody: "More devlog posts can now reuse this pattern for clips, screenshots, and mixed media updates."
utilityText: "This post is a focused asset-loading check for the Astro devlog flow."
finePrint: "Space Rocks! is an active development project. All media, features, names, and content are work in progress and subject to change before final release."
heroMediaKind: "youtube"
heroImages: []
heroYoutubeUrl: "https://www.youtube.com/watch?v=OwgTiV9-jAA"
heroMediaAlt: "Testing Media Assets hero YouTube video"
articleMediaKind: "images"
articleImages:
  - "/media/devlog/0003-testing-media-assets/asteroid1.png"
  - "/media/devlog/0003-testing-media-assets/asteroid2.png"
  - "/media/devlog/0003-testing-media-assets/asteroid3.png"
  - "/media/devlog/0003-testing-media-assets/asteroid4.png"
  - "/media/devlog/0003-testing-media-assets/asteroid5.png"
  - "/media/devlog/0003-testing-media-assets/asteroid6.png"
  - "/media/devlog/0003-testing-media-assets/asteroid7.png"
  - "/media/devlog/0003-testing-media-assets/asteroid8.png"
articleYoutubeUrl: ""
articleMediaAlt: "Asteroid media gallery for devlog entry 003"
---

This entry is a real-media smoke test for the devlog pipeline. The top frame uses the provided YouTube link, and the article frame cycles through the full eight-image asteroid set copied into the public media directory for route-safe static serving.
