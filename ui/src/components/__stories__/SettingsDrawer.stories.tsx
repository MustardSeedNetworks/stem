import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { defaultRFC2544Config, type RFC2544Config } from '../RFC2544ConfigForm';
import { defaultRFC2889Config, type RFC2889Config } from '../RFC2889ConfigForm';
import { defaultRFC6349Config, type RFC6349Config } from '../RFC6349ConfigForm';
import { SettingsDrawer } from '../SettingsDrawer';
import { defaultTrafficGenConfig, type TrafficGenConfig } from '../TrafficGenConfigForm';
import { defaultTSNConfig, type TSNConfig } from '../TSNConfigForm';
import { defaultY1564Config, type Y1564Config } from '../Y1564ConfigForm';
import { defaultY1731Config, type Y1731Config } from '../Y1731ConfigForm';

const meta: Meta<typeof SettingsDrawer> = {
  title: 'Components/SettingsDrawer',
  component: SettingsDrawer,
  parameters: { layout: 'fullscreen' },
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof SettingsDrawer>;

export const Default: Story = {
  render: () => {
    const [selectedTests, setSelectedTests] = useState<string[]>([
      'rfc2544_throughput',
      'y1564_config',
      'rfc2889_forwarding',
      'rfc6349_throughput',
    ]);
    const [rfc2544Config, setRFC2544Config] = useState<RFC2544Config>(defaultRFC2544Config);
    const [rfc2889Config, setRFC2889Config] = useState<RFC2889Config>(defaultRFC2889Config);
    const [rfc6349Config, setRFC6349Config] = useState<RFC6349Config>(defaultRFC6349Config);
    const [y1564Config, setY1564Config] = useState<Y1564Config>(defaultY1564Config);
    const [y1731Config, setY1731Config] = useState<Y1731Config>(defaultY1731Config);
    const [tsnConfig, setTSNConfig] = useState<TSNConfig>(defaultTSNConfig);
    const [trafficGenConfig, setTrafficGenConfig] =
      useState<TrafficGenConfig>(defaultTrafficGenConfig);

    return (
      <SettingsDrawer
        isOpen={true}
        onClose={() => {}}
        selectedTests={selectedTests}
        setSelectedTests={setSelectedTests}
        rfc2544Config={rfc2544Config}
        setRFC2544Config={setRFC2544Config}
        rfc2889Config={rfc2889Config}
        setRFC2889Config={setRFC2889Config}
        rfc6349Config={rfc6349Config}
        setRFC6349Config={setRFC6349Config}
        y1564Config={y1564Config}
        setY1564Config={setY1564Config}
        y1731Config={y1731Config}
        setY1731Config={setY1731Config}
        tsnConfig={tsnConfig}
        setTSNConfig={setTSNConfig}
        trafficGenConfig={trafficGenConfig}
        setTrafficGenConfig={setTrafficGenConfig}
      />
    );
  },
};
