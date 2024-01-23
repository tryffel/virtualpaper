import { useRecordContext } from "react-admin";
import {
  Grid,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import { get } from "lodash";

export const ServerStatistics = () => {
  const record = useRecordContext();
  return (
    <TableContainer style={{ maxWidth: "500px" }}>
      <Typography variant="h6">Server</Typography>
      <Table size={"small"}>
        <TableHead>
          <TableRow>
            <TableCell></TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow>
            <TableCell>Uptime</TableCell>
            <TableCell>{get(record, "uptime")}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Load</TableCell>
            <TableCell>{get(record, "server_load")}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Total documents</TableCell>
            <TableCell>{get(record, "documents_total")}</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>Total space used</TableCell>
            <TableCell>{get(record, "documents_total_size_string")}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </TableContainer>
  );
};
export const ProcessingStatistics = () => {
  const record = useRecordContext();
  return (
    <Grid container spacing={1}>
      <Grid item xs={12}>
        <Typography variant="h6">Document processing</Typography>
      </Grid>
      <Grid item xs={12}>
        <TableContainer style={{ maxWidth: "500px" }}>
          <Table size={"small"}>
            <TableHead>
              <TableRow>
                <TableCell></TableCell>
                <TableCell></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              <TableRow>
                <TableCell>Queued</TableCell>
                <TableCell>{get(record, "documents_queued")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Processed today</TableCell>
                <TableCell>
                  {get(record, "documents_processed_today")}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Processed past week</TableCell>
                <TableCell>
                  {get(record, "documents_processed_past_week")}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Processed past month</TableCell>
                <TableCell>
                  {get(record, "documents_processed_past_month")}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Grid>
    </Grid>
  );
};

export const InstallationInfo = () => {
  const record = useRecordContext();
  return (
    <Grid container spacing={1}>
      <Grid item xs={12}>
        <Typography variant="h6">Installation</Typography>
      </Grid>
      <Grid item xs={12}>
        <TableContainer style={{ maxWidth: "500px" }}>
          <Table size={"small"}>
            <TableHead>
              <TableRow>
                <TableCell></TableCell>
                <TableCell></TableCell>
              </TableRow>
            </TableHead>

            <TableBody>
              <TableRow>
                <TableCell>Imagemagick version</TableCell>
                <TableCell>{get(record, "imagemagick_version")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Tesseract version</TableCell>
                <TableCell>{get(record, "tesseract_version")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Poppler installed</TableCell>
                <TableCell>{get(record, "poppler_installed")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Pandoc installed</TableCell>
                <TableCell>{get(record, "pandoc_installed")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Number of CPUs</TableCell>
                <TableCell>{get(record, "number_cpus")}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Grid>
    </Grid>
  );
};

export const BuildInfo = () => {
  const record = useRecordContext();

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Typography variant="h6">General information</Typography>
      </Grid>
      <Grid item xs={12}>
        <TableContainer style={{ maxWidth: "500px" }}>
          <Table size={"small"}>
            <TableHead>
              <TableRow>
                <TableCell></TableCell>
                <TableCell></TableCell>
              </TableRow>
            </TableHead>

            <TableBody>
              <TableRow>
                <TableCell>Virtualpaper</TableCell>
                <TableCell>
                  <a href={"https://github.com/tryffel/virtualpaper"}>
                    github.com/tryffel/virtualpaper
                  </a>
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Documentation</TableCell>
                <TableCell>
                  <a href={"https://virtualpaper.tryffel.net"}>
                    virtualpaper.tryffel.net
                  </a>
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Version</TableCell>
                <TableCell>{get(record, "version")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Git commit</TableCell>
                <TableCell>{get(record, "commit")}</TableCell>
              </TableRow>
              <TableRow>
                <TableCell>Go version</TableCell>
                <TableCell>{get(record, "go_version")}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Grid>
    </Grid>
  );
};
