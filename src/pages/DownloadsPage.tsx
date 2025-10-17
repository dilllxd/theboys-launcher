import React from 'react';
import styled from 'styled-components';
import DownloadManager from '../components/DownloadManager';
import { Layout } from '../components/layout/Layout';

const DownloadsPageContainer = styled.div`
  height: 100%;
  display: flex;
  flex-direction: column;
`;

const PageHeader = styled.div`
  padding: var(--spacing-lg) var(--spacing-xl);
  border-bottom: 1px solid var(--color-border);
`;

const PageTitle = styled.h1`
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
`;

const PageDescription = styled.p`
  color: var(--color-text-secondary);
  margin: var(--spacing-sm) 0 0 0;
  font-size: var(--font-size-base);
`;

const Content = styled.div`
  flex: 1;
  padding: var(--spacing-lg) var(--spacing-xl);
  overflow: hidden;
  display: flex;
  flex-direction: column;
`;

export const DownloadsPage: React.FC = () => {
  return (
    <Layout>
      <DownloadsPageContainer>
        <PageHeader>
          <PageTitle>Downloads</PageTitle>
          <PageDescription>
            Manage your ongoing and completed downloads, including modpacks, tools, and resources.
          </PageDescription>
        </PageHeader>
        <Content>
          <DownloadManager />
        </Content>
      </DownloadsPageContainer>
    </Layout>
  );
};

export default DownloadsPage;