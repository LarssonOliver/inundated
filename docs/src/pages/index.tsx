import type {CSSProperties, ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

type Feature = {
  label: string;
  color: string;
  description: string;
};

const features: Feature[] = [
  {
    label: 'Time tracking',
    color: 'var(--nord8)',
    description:
      'Log entries with a start and end time, then group them with tags — no running timer to forget about.',
  },
  {
    label: 'Projects & budgets',
    color: 'var(--nord14)',
    description:
      'Roll tags up into projects with an optional time budget, without touching any historical entries.',
  },
  {
    label: 'Self-hosted',
    color: 'var(--nord13)',
    description:
      'One binary, one Postgres database. Runs standalone with no login, or behind your own OIDC provider.',
  },
];

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  const logo = useBaseUrl('img/logo.svg');
  const screenshot = useBaseUrl('img/screenshot.jpeg');
  return (
    <header className={styles.heroBanner}>
      <div className={clsx('container', styles.heroInner)}>
        <div className={styles.heroContent}>
          <div className={styles.brand}>
            <img src={logo} alt="" className={styles.brandLogo} />
            <span className={styles.brandName}>{siteConfig.title}</span>
          </div>
          <Heading as="h1" className={styles.heroTitle}>
            Time tracking, self-hosted.
          </Heading>
          <p className={styles.heroSubtitle}>
            Inundated is a lightweight app for logging time against tags and
            projects. One binary, one Postgres database, no subscription —
            runs standalone or behind your own OIDC login.
          </p>
          <div className={styles.buttons}>
            <Link
              className="button button--primary button--lg"
              to="/docs/intro">
              Read the docs
            </Link>
            <Link
              className="button button--secondary button--lg"
              to="https://github.com/LarssonOliver/inundated">
              View on GitHub
            </Link>
          </div>
        </div>
        <div className={styles.heroMedia}>
          <div className={styles.screenshotFrame}>
            <img
              src={screenshot}
              alt="The Inundated timesheet, showing logged time entries grouped by day"
              className={styles.screenshot}
            />
          </div>
        </div>
      </div>
    </header>
  );
}

function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className={clsx('container', styles.featuresGrid)}>
        {features.map((feature) => (
          <div key={feature.label} className={styles.feature}>
            <span
              className={styles.featurePill}
              style={{'--feature-color': feature.color} as CSSProperties}>
              {feature.label}
            </span>
            <p className={styles.featureDescription}>{feature.description}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout title={siteConfig.title} description={siteConfig.tagline}>
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
