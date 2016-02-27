package btcloud

import (	
	"fmt"
	"io/ioutil"
	"time"
	
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/log"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type Driver struct {
	*drivers.BaseDriver
	VmId			string
	Endpoint		string
	APIKey			string
	SecretKey		string
	VerifySSL		bool
	Zone			string
	Template		string
	ServiceOffering		string
	IPAddress		string	
	SSHKeypair		string
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag {
		mcnflag.StringFlag{
			EnvVar: "BTCLOUD_ENDPOINT",
			Name:   "btcloud-endpoint",
			Usage:  "BT Cloud Compute API endpoint (URL)",
		},
		mcnflag.StringFlag{
			EnvVar: "BTCLOUD_API_KEY",
			Name:   "btcloud-api-key",
			Usage:  "BT Cloud Compute API key",
		},
		mcnflag.StringFlag{
			EnvVar: "BTCLOUD_SECRET_KEY",
			Name:   "btcloud-secret-key",
			Usage:  "BT Cloud Compute secret key",
		},
		mcnflag.BoolFlag{
			EnvVar: "BTCLOUD_SSL",
			Name:   "btcloud-ssl",
			Usage:  "Verify BT Cloud Compute SSL",
		},
		mcnflag.StringFlag{
			EnvVar: "BTCLOUD_ZONE",
			Name:   "btcloud-zone",
			Usage:  "BT Cloud Compute availability zone",
		},
		mcnflag.StringFlag{
			EnvVar: "BTCLOUD_TEMPLATE",
			Name:   "btcloud-template",
			Usage:  "BT Cloud Compute template",
		},
		mcnflag.StringFlag{
			EnvVar: "BTCLOUD_SERVICE_OFFERING",
			Name:   "btcloud-service-offering",
			Usage:  "BT Cloud Compute service offering",
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Endpoint = flags.String("btcloud-endpoint")
	d.APIKey = flags.String("btcloud-api-key")
	d.SecretKey = flags.String("btcloud-secret-key")
	d.VerifySSL	= flags.Bool("btcloud-ssl")
	d.Zone = flags.String("btcloud-zone")
	d.Template = flags.String("btcloud-template")
	d.ServiceOffering = flags.String("btcloud-service-offering")
		
	return d.checkConfig()
}

func (d *Driver) checkConfig() error {
	if d.Endpoint == "" {
		d.Endpoint = "https://cloud.bt.com/client/api"
	}
	if d.APIKey == "" {
		return fmt.Errorf("Please specify an API key (--btcloud-api-key).")
	}
	if d.SecretKey == "" {
		return fmt.Errorf("Please specify a secret key (--btcloud-secret-key).")
	}
	if d.Zone == "" {
		return fmt.Errorf("Please specify an availability zone (--btcloud-zone).")
	}
	if d.Template == "" {
		return fmt.Errorf("Please specify an Ubuntu template (--btcloud-template).")
	}
	if d.ServiceOffering == "" {
		return fmt.Errorf("Please specify a service offering (--btcloud-service-offering).")
	}

	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) DriverName() string {
	return "btcloud"
}

func (d *Driver) Create() error {
	
	// <TO DO = Automate and fix mapping>
	zoneId := d.Zone
	templateId := d.Template
	serviceOfferingId := d.ServiceOffering
	// </TO DO>
	
	client := cloudstack.NewAsyncClient(d.Endpoint, d.APIKey, d.SecretKey, d.VerifySSL)
	randomKey := mcnutils.TruncateID(mcnutils.GenerateRandomID())
	d.SSHKeypair = fmt.Sprintf("%s-%s", d.MachineName, randomKey)
	kpService := cloudstack.NewSSHService(client)
	kpParams := kpService.NewCreateSSHKeyPairParams(d.SSHKeypair)
	kpResp, err := kpService.CreateSSHKeyPair(kpParams)
	if err != nil {
		return err
	}
  err = ioutil.WriteFile(d.GetSSHKeyPath(), []byte(kpResp.Privatekey), 0600)
  if err != nil {
		return err
	}
	log.Debugf("SSH Keypair ID = %s", d.SSHKeypair)

	vmService := cloudstack.NewVirtualMachineService(client)
	vmParams := vmService.NewDeployVirtualMachineParams(serviceOfferingId, templateId, zoneId)
	vmParams.SetKeypair(d.SSHKeypair)
	vmParams.SetName(d.MachineName)
	vmParams.SetDisplayname("Docker Machine VM")
	vmResp, err := vmService.DeployVirtualMachine(vmParams)
	if err != nil {
		 return err
	}
	d.VmId = vmResp.Id
	vm, err := d.getVmByID(d.VmId)
	if err != nil {
		 return err
	}
	time.Sleep(3000) // Give CloudStack some time
	networkId := vm.Nic[0].Networkid
	log.Debugf("VM / Network ID = %s / %s", d.VmId, networkId)

	ipService := cloudstack.NewAddressService(client)
	ipParams := ipService.NewListPublicIpAddressesParams()
	ipParams.SetAssociatednetworkid(networkId)
	ipResp, err := ipService.ListPublicIpAddresses(ipParams)
	if err != nil {
		 return err
	}
  ipAddressId := ipResp.PublicIpAddresses[0].Id
  d.IPAddress = ipResp.PublicIpAddresses[0].Ipaddress
  log.Debugf("IP address = %s", d.IPAddress)

	fwService := cloudstack.NewFirewallService(client)
	fwParams := fwService.NewCreatePortForwardingRuleParams(ipAddressId, 22, "TCP", 22, d.VmId)
	_, err = fwService.CreatePortForwardingRule(fwParams)
	fwParams = fwService.NewCreatePortForwardingRuleParams(ipAddressId, 2376, "TCP", 2376, d.VmId)
	_, err = fwService.CreatePortForwardingRule(fwParams)
	fwParams = fwService.NewCreatePortForwardingRuleParams(ipAddressId, 3376, "TCP", 3376, d.VmId)
	_, err = fwService.CreatePortForwardingRule(fwParams)
	return err
}

func (d *Driver) GetState() (state.State, error) {
	vm, err := d.getVmByID(d.VmId)
	if err != nil {
		return state.Error, err
	}	
	vmState := state.None
	switch vm.State {
	case "Starting":
		vmState = state.Starting
	case "Running", "Stopping":
		vmState = state.Running
	case "Stopped", "Destroyed", "Expunging", "Shutdowned":
		vmState = state.Stopped
	default:
		return state.Error, nil 
	}
	return vmState, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return "root"
}

func (d *Driver) GetIP() (string, error) {	
	return d.IPAddress, nil
}

func (d *Driver) GetURL() (string, error) {
  ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("tcp://%s:2376", ip)
	return url, nil
}

func (d *Driver) Start() error {
	log.Debugf("Starting virtual machine %s", d.VmId)
	client := cloudstack.NewClient(d.Endpoint, d.APIKey, d.SecretKey, d.VerifySSL)
	vmService := cloudstack.NewVirtualMachineService(client)
	vmParams := vmService.NewStartVirtualMachineParams(d.VmId)
	_, err := vmService.StartVirtualMachine(vmParams)

	return err
}

func (d *Driver) Restart() error {
	log.Debugf("Restarting virtual machine %s", d.VmId)
	client := cloudstack.NewClient(d.Endpoint, d.APIKey, d.SecretKey, d.VerifySSL)
	vmService := cloudstack.NewVirtualMachineService(client)
	vmParams := vmService.NewRebootVirtualMachineParams(d.VmId)
	_, err := vmService.RebootVirtualMachine(vmParams)
	
	return err
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Stop() error {
	log.Debugf("Stopping virtual machine %s", d.VmId)
	client := cloudstack.NewClient(d.Endpoint, d.APIKey, d.SecretKey, d.VerifySSL)
	vmService := cloudstack.NewVirtualMachineService(client)
	vmParams := vmService.NewStopVirtualMachineParams(d.VmId)
	_, err := vmService.StopVirtualMachine(vmParams)

	return err
}

func (d *Driver) Remove() error {
	log.Debugf("Destroying virtual machine %s", d.VmId)
	client := cloudstack.NewClient(d.Endpoint, d.APIKey, d.SecretKey, d.VerifySSL)
	vmService := cloudstack.NewVirtualMachineService(client)
	vmParams := vmService.NewDestroyVirtualMachineParams(d.VmId)
	vmParams.SetExpunge(true)
	_, err := vmService.DestroyVirtualMachine(vmParams)
	if err != nil {
		 return err
	}
	
	log.Debugf("Destroying SSH keypair")
	kpService := cloudstack.NewSSHService(client)
	kpParams := kpService.NewDeleteSSHKeyPairParams(d.SSHKeypair)
	_, err = kpService.DeleteSSHKeyPair(kpParams)

	return err
}

func (d *Driver) getVmByID(id string) (vm *cloudstack.VirtualMachine, err error) {
	client := cloudstack.NewClient(d.Endpoint, d.APIKey, d.SecretKey, d.VerifySSL)
	service := cloudstack.NewVirtualMachineService(client)
	vm, _, err = service.GetVirtualMachineByID(id)
		
	return vm, err
}
