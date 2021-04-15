export interface StatusModel {
  healthy: boolean;
  name: string;
  support: string[];
  version: string;
  buildTime: number;
  buildDate: Date;
}

export interface StatusDetailedModel extends StatusModel {
  trackers: Tracker[]
}

export interface Tracker {
  Name: string
  CurrentProblems: Map<string, string>
  CurrentWarnings: Map<string, string>
}
