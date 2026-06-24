import type { GetStaticProps, InferGetStaticPropsType } from "next";
import {
  PlasmicComponent,
  PlasmicRootProvider,
  type ComponentRenderData,
} from "@plasmicapp/loader-nextjs";

import { PLASMIC } from "@/src/plasmic/plasmic-init";

const HOMEPAGE_COMPONENT = "Homepage";

type HomePageProps = {
  plasmicData: ComponentRenderData;
};

export const getStaticProps: GetStaticProps<HomePageProps> = async () => {
  const plasmicData = await PLASMIC.fetchComponentData(HOMEPAGE_COMPONENT);

  return {
    props: {
      plasmicData,
    },
  };
};

export default function HomePage({
  plasmicData,
}: InferGetStaticPropsType<typeof getStaticProps>) {
  return (
    <PlasmicRootProvider loader={PLASMIC} prefetchedData={plasmicData}>
      <PlasmicComponent component={HOMEPAGE_COMPONENT} />
    </PlasmicRootProvider>
  );
}
